/**
 * Express 애플리케이션에 Prometheus 메트릭 통합 예시
 *
 * 설치:
 * npm install express prom-client
 *
 * 실행:
 * node nodejs-express-metrics.js
 *
 * 메트릭 확인:
 * http://localhost:3000/metrics
 */

const express = require('express');
const promClient = require('prom-client');

const app = express();
app.use(express.json());

// Prometheus 기본 메트릭 활성화 (CPU, 메모리 등)
promClient.collectDefaultMetrics({ timeout: 5000 });

// 커스텀 메트릭 정의
const httpRequestCounter = new promClient.Counter({
  name: 'http_requests_total',
  help: 'Total number of HTTP requests',
  labelNames: ['method', 'route', 'status']
});

const httpRequestDuration = new promClient.Histogram({
  name: 'http_request_duration_seconds',
  help: 'Duration of HTTP requests in seconds',
  labelNames: ['method', 'route'],
  buckets: [0.001, 0.01, 0.1, 0.5, 1, 2, 5]
});

const activeRequests = new promClient.Gauge({
  name: 'http_requests_active',
  help: 'Number of active HTTP requests',
  labelNames: ['method', 'route']
});

// 비즈니스 메트릭
const businessRevenue = new promClient.Counter({
  name: 'business_revenue_total',
  help: 'Total revenue',
  labelNames: ['currency']
});

const businessOrders = new promClient.Counter({
  name: 'business_orders_total',
  help: 'Total orders',
  labelNames: ['product_type', 'status']
});

const activeUsers = new promClient.Gauge({
  name: 'active_users',
  help: 'Number of currently active users'
});

const dbQueryDuration = new promClient.Histogram({
  name: 'db_query_duration_seconds',
  help: 'Database query duration',
  labelNames: ['query_type'],
  buckets: [0.001, 0.01, 0.1, 0.5, 1]
});

// 메트릭 수집 미들웨어
app.use((req, res, next) => {
  const start = Date.now();
  const route = req.route?.path || req.path;

  // 활성 요청 증가
  activeRequests.inc({ method: req.method, route });

  // 응답 완료 시
  res.on('finish', () => {
    const duration = (Date.now() - start) / 1000;

    // 요청 카운트 증가
    httpRequestCounter.inc({
      method: req.method,
      route,
      status: res.statusCode
    });

    // 요청 지속시간 기록
    httpRequestDuration.observe({
      method: req.method,
      route
    }, duration);

    // 활성 요청 감소
    activeRequests.dec({ method: req.method, route });
  });

  next();
});

// 데이터베이스 쿼리 시뮬레이션
function simulateDbQuery(queryType = 'select') {
  return new Promise((resolve) => {
    const end = dbQueryDuration.startTimer({ query_type: queryType });

    // 실제로는 DB 쿼리 실행
    setTimeout(() => {
      end();
      resolve();
    }, Math.random() * 100);
  });
}

// API 엔드포인트
app.get('/', (req, res) => {
  res.json({
    message: 'Express Prometheus Metrics Example',
    metrics_endpoint: '/metrics'
  });
});

app.get('/api/users', async (req, res) => {
  activeUsers.inc();

  try {
    await simulateDbQuery('select');

    res.json({
      users: [
        { id: 1, name: 'User 1' },
        { id: 2, name: 'User 2' }
      ]
    });
  } finally {
    activeUsers.dec();
  }
});

app.post('/api/order', async (req, res) => {
  const { product_type = 'unknown', amount = 0, currency = 'USD' } = req.body;

  // 주문 처리
  await simulateDbQuery('insert');

  // 성공/실패 랜덤 시뮬레이션 (90% 성공률)
  const success = Math.random() > 0.1;

  if (success) {
    // 비즈니스 메트릭 기록
    businessOrders.inc({ product_type, status: 'success' });
    businessRevenue.inc({ currency }, amount);

    res.json({
      status: 'success',
      order_id: Math.floor(Math.random() * 9000) + 1000
    });
  } else {
    businessOrders.inc({ product_type, status: 'failed' });

    res.status(500).json({
      status: 'failed',
      error: 'Payment failed'
    });
  }
});

app.get('/api/slow', async (req, res) => {
  // 느린 엔드포인트 시뮬레이션
  await new Promise(resolve => setTimeout(resolve, Math.random() * 2000 + 1000));
  res.json({ message: 'This is slow' });
});

app.get('/api/error', (req, res) => {
  // 에러 시뮬레이션
  if (Math.random() > 0.5) {
    throw new Error('Random error');
  }
  res.json({ message: 'Success' });
});

// Prometheus 메트릭 엔드포인트
app.get('/metrics', async (req, res) => {
  res.set('Content-Type', promClient.register.contentType);
  res.end(await promClient.register.metrics());
});

// 에러 핸들러
app.use((err, req, res, next) => {
  console.error(err.stack);
  res.status(500).json({ error: err.message });
});

// 서버 시작
const PORT = process.env.PORT || 3000;
app.listen(PORT, () => {
  console.log(`Express app listening at http://localhost:${PORT}`);
  console.log(`Metrics available at http://localhost:${PORT}/metrics`);
});
