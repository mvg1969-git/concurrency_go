package server_test

import (
	"context"
	"net"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"

	"mv_db/internal/config"
	"mv_db/network"
)

// Мок базы данных, который заблокирует обработку, пока мы не разрешим в тесте
type mockDb struct {
	mu           sync.Mutex
	totalHandled int
	releaseChan  chan struct{} // канал для управления разблокировкой, иммитрует долгий запрос к бд
}

func (m *mockDb) HandleQuery(ctx context.Context, queryStr string) string {
	m.mu.Lock()
	m.totalHandled++
	m.mu.Unlock()

	// Зависаем здесь, имитируя обработку, пока тест не даст команду освободить семафор
	if m.releaseChan != nil {
		<-m.releaseChan
	}

	return "ok"
}

var _ = Describe("Server Semaphore Integration", func() {
	var (
		cfg        *config.Config
		srv        *network.TCPServer
		logger     *zap.Logger
		ctx        context.Context
		cancel     context.CancelFunc
		db         *mockDb
		srvAddress string
	)

	BeforeEach(func() {
		logger = zap.NewNop() // отключаем логи
		ctx, cancel = context.WithCancel(context.Background())

		db = &mockDb{
			releaseChan: make(chan struct{}), // Инициализируем канал блокировки
		}

		cfg = &config.Config{
			Network: config.NetworkConfig{
				Address:        "127.0.0.1:3225",
				MaxConnections: 2, // Сервер пустит только 2 горутины
				MaxMessageSize: 1024,
				IdleTimeout:    config.Duration(5 * time.Second),
			},
		}

		var options []network.TCPServerOption
		if cfg.Network.MaxConnections != 0 {
			options = append(options, network.WithServerMaxConnectionsNumber(uint(cfg.Network.MaxConnections)))
		}

		srv = network.NewTCPServer(logger, cfg.Network.Address, db, options...)
		srvAddress = cfg.Network.Address

		go func() {
			defer GinkgoRecover()
			_ = srv.Start(ctx)
		}()

		// Ожидаем старта сервера
		time.Sleep(50 * time.Millisecond)
	})

	AfterEach(func() {
		cancel()
		time.Sleep(50 * time.Millisecond)
	})

	Context("Когда количество клиентов превышает MaxConnections", func() {
		It("должен обрабатывать не более MaxConnections соединений, блокируя остальные в семафоре", func() {
			const totalClients = 3
			var wg sync.WaitGroup

			// Запускаем 3 клиентов параллельно
			for i := 0; i < totalClients; i++ {
				wg.Add(1)
				go func() {
					defer GinkgoRecover()
					defer wg.Done()

					conn, err := net.DialTimeout("tcp", srvAddress, 1*time.Second)
					Expect(err).To(Not(HaveOccurred()))
					defer conn.Close()

					// Отправляем запрос с переносом строки
					_, err = conn.Write([]byte("select_test\n"))
					Expect(err).To(Not(HaveOccurred()))

					// Ждем ответа "ok\n"
					buf := make([]byte, 64)
					_, _ = conn.Read(buf)
				}()
			}

			// ШАГ 1: Даем горутинам клиентов долететь до сервера.
			// Ждем (Eventually), пока ровно 2 клиента гарантированно зайдут в базу данных.
			Eventually(func() int {
				db.mu.Lock()
				defer db.mu.Unlock()
				return db.totalHandled
			}, "500ms", "10ms").Should(Equal(2),
				"Сервер не смог принять даже первые 2 разрешенных соединения за полсекунды")

			// ШАГ 2: Теперь проверяем устойчивость (Consistently).
			// В течение следующих 150мс счетчик должен СТРОГО оставаться равным 2 (3-й клиент заблокирован).
			Consistently(func() int {
				db.mu.Lock()
				defer db.mu.Unlock()
				return db.totalHandled
			}, "150ms", "10ms").Should(Equal(2),
				"Семафор пропустил 3-го клиента, лимит соединений нарушен!")

			// ШАГ 3: Разблокируем канал в моке БД, позволяя серверу дообрабатывать запросы
			close(db.releaseChan)

			// Ждем завершения работы всех сетевых горутин клиентов
			wg.Wait()

			// ШАГ 4: Финальная проверка — после открытия шлюза 3-й клиент тоже успешно обработался
			db.mu.Lock()
			Expect(db.totalHandled).To(Equal(3))
			db.mu.Unlock()
		})
	})
})
