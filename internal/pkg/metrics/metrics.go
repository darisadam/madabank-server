package metrics

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP Metrics
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "madabank_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "madabank_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	// Transaction Metrics
	TransactionsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "madabank_transactions_total",
			Help: "Total number of transactions",
		},
		[]string{"type", "status"},
	)

	TransactionAmount = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "madabank_transaction_amount",
			Help:    "Transaction amounts",
			Buckets: []float64{10, 50, 100, 500, 1000, 5000, 10000, 50000, 100000},
		},
		[]string{"type", "currency"},
	)

	TransactionDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "madabank_transaction_duration_seconds",
			Help:    "Transaction processing duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 2, 5},
		},
		[]string{"type"},
	)

	TransactionErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "madabank_transaction_errors_total",
			Help: "Total number of transaction errors",
		},
		[]string{"type", "error_type"},
	)

	// Account Metrics
	AccountsTotal = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "madabank_accounts_total",
			Help: "Total number of active accounts",
		},
		[]string{"type"},
	)

	AccountBalance = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "madabank_account_balance",
			Help: "Current account balances",
		},
		[]string{"account_id", "type", "currency"},
	)

	TotalBalanceByType = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "madabank_total_balance_by_type",
			Help: "Total balance by account type",
		},
		[]string{"type", "currency"},
	)

	// User Metrics
	UsersTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "madabank_users_total",
			Help: "Total number of registered users",
		},
	)

	ActiveUsersTotal = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "madabank_active_users_total",
			Help: "Total number of active users",
		},
	)

	// Authentication Metrics
	AuthAttemptsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "madabank_auth_attempts_total",
			Help: "Total number of authentication attempts",
		},
		[]string{"status"},
	)

	AuthTokensGenerated = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "madabank_auth_tokens_generated_total",
			Help: "Total number of JWT tokens generated",
		},
	)

	// Database Metrics
	DBConnectionsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "madabank_db_connections_active",
			Help: "Number of active database connections",
		},
	)

	DBQueriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "madabank_db_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"operation", "table"},
	)

	DBQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "madabank_db_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1},
		},
		[]string{"operation", "table"},
	)

	// System Metrics
	SystemInfo = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "madabank_system_info",
			Help: "System information",
		},
		[]string{"version", "commit_sha", "go_version"},
	)
)

// RecordHTTPRequest records HTTP request metrics
func RecordHTTPRequest(method, endpoint string, status int, duration float64) {
	HTTPRequestsTotal.WithLabelValues(method, endpoint, strconv.Itoa(status)).Inc()
	HTTPRequestDuration.WithLabelValues(method, endpoint).Observe(duration)
}

// RecordTransaction records transaction metrics
func RecordTransaction(txnType, status string, amount float64, currency string, duration float64) {
	TransactionsTotal.WithLabelValues(txnType, status).Inc()
	TransactionAmount.WithLabelValues(txnType, currency).Observe(amount)
	TransactionDuration.WithLabelValues(txnType).Observe(duration)
}

// RecordTransactionError records transaction error
func RecordTransactionError(txnType, errorType string) {
	TransactionErrors.WithLabelValues(txnType, errorType).Inc()
}

// RecordAuthAttempt records authentication attempt
func RecordAuthAttempt(success bool) {
	status := "failed"
	if success {
		status = "success"
	}
	AuthAttemptsTotal.WithLabelValues(status).Inc()
}

// RecordAuthTokenGenerated records JWT token generation
func RecordAuthTokenGenerated() {
	AuthTokensGenerated.Inc()
}

// UpdateAccountMetrics updates account-related metrics
func UpdateAccountMetrics(accountType string, count int) {
	AccountsTotal.WithLabelValues(accountType).Set(float64(count))
}

// UpdateUserMetrics updates user-related metrics
func UpdateUserMetrics(total, active int) {
	UsersTotal.Set(float64(total))
	ActiveUsersTotal.Set(float64(active))
}

// RecordDBQuery records database query metrics
func RecordDBQuery(operation, table string, duration float64) {
	DBQueriesTotal.WithLabelValues(operation, table).Inc()
	DBQueryDuration.WithLabelValues(operation, table).Observe(duration)
}

// SetSystemInfo sets system information metrics
func SetSystemInfo(version, commitSHA, goVersion string) {
	SystemInfo.WithLabelValues(version, commitSHA, goVersion).Set(1)
}
