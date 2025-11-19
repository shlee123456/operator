#!/usr/bin/env python3
"""
Flask 애플리케이션에 Prometheus 메트릭 통합 예시

실행 방법:
pip install flask prometheus-client
python python-flask-metrics.py

메트릭 확인:
http://localhost:8000/metrics
"""

from flask import Flask, request, jsonify
from prometheus_client import Counter, Histogram, Gauge, start_http_server, generate_latest
from functools import wraps
import time
import random

app = Flask(__name__)

# 메트릭 정의
REQUEST_COUNT = Counter(
    'http_requests_total',
    'Total HTTP requests',
    ['method', 'endpoint', 'status']
)

REQUEST_DURATION = Histogram(
    'http_request_duration_seconds',
    'HTTP request latency in seconds',
    ['method', 'endpoint']
)

ACTIVE_REQUESTS = Gauge(
    'http_requests_active',
    'Number of active HTTP requests',
    ['method', 'endpoint']
)

BUSINESS_REVENUE = Counter(
    'business_revenue_total',
    'Total revenue',
    ['currency']
)

BUSINESS_ORDERS = Counter(
    'business_orders_total',
    'Total orders',
    ['product_type', 'status']
)

ACTIVE_USERS = Gauge(
    'active_users',
    'Number of currently active users'
)

DB_QUERY_DURATION = Histogram(
    'db_query_duration_seconds',
    'Database query duration',
    ['query_type']
)


# 메트릭 데코레이터
def track_metrics(f):
    @wraps(f)
    def decorated_function(*args, **kwargs):
        endpoint = request.endpoint or 'unknown'
        method = request.method

        # 활성 요청 증가
        ACTIVE_REQUESTS.labels(method=method, endpoint=endpoint).inc()

        # 시작 시간
        start_time = time.time()

        try:
            # 함수 실행
            response = f(*args, **kwargs)
            status = 200

            return response

        except Exception as e:
            status = 500
            raise e

        finally:
            # 요청 지속시간 기록
            duration = time.time() - start_time
            REQUEST_DURATION.labels(
                method=method,
                endpoint=endpoint
            ).observe(duration)

            # 요청 카운트 증가
            REQUEST_COUNT.labels(
                method=method,
                endpoint=endpoint,
                status=status
            ).inc()

            # 활성 요청 감소
            ACTIVE_REQUESTS.labels(method=method, endpoint=endpoint).dec()

    return decorated_function


# 비즈니스 로직 시뮬레이션
def simulate_db_query(query_type='select'):
    """데이터베이스 쿼리 시뮬레이션"""
    with DB_QUERY_DURATION.labels(query_type=query_type).time():
        # 실제로는 DB 쿼리 실행
        time.sleep(random.uniform(0.01, 0.1))


# API 엔드포인트
@app.route('/')
@track_metrics
def index():
    return jsonify({
        'message': 'Flask Prometheus Metrics Example',
        'metrics_endpoint': '/metrics'
    })


@app.route('/api/users')
@track_metrics
def get_users():
    ACTIVE_USERS.inc()
    try:
        simulate_db_query('select')
        return jsonify({
            'users': [
                {'id': 1, 'name': 'User 1'},
                {'id': 2, 'name': 'User 2'}
            ]
        })
    finally:
        ACTIVE_USERS.dec()


@app.route('/api/order', methods=['POST'])
@track_metrics
def create_order():
    """주문 생성 (비즈니스 메트릭 예시)"""
    data = request.get_json() or {}

    product_type = data.get('product_type', 'unknown')
    amount = data.get('amount', 0)
    currency = data.get('currency', 'USD')

    # 주문 처리
    simulate_db_query('insert')

    # 성공/실패 랜덤 시뮬레이션
    success = random.random() > 0.1  # 90% 성공률

    if success:
        # 비즈니스 메트릭 기록
        BUSINESS_ORDERS.labels(
            product_type=product_type,
            status='success'
        ).inc()

        BUSINESS_REVENUE.labels(currency=currency).inc(amount)

        return jsonify({
            'status': 'success',
            'order_id': random.randint(1000, 9999)
        })
    else:
        BUSINESS_ORDERS.labels(
            product_type=product_type,
            status='failed'
        ).inc()

        return jsonify({
            'status': 'failed',
            'error': 'Payment failed'
        }), 500


@app.route('/api/slow')
@track_metrics
def slow_endpoint():
    """느린 엔드포인트 시뮬레이션"""
    time.sleep(random.uniform(1, 3))
    return jsonify({'message': 'This is slow'})


@app.route('/api/error')
@track_metrics
def error_endpoint():
    """에러 시뮬레이션"""
    if random.random() > 0.5:
        raise Exception("Random error")
    return jsonify({'message': 'Success'})


@app.route('/metrics')
def metrics():
    """Prometheus 메트릭 엔드포인트"""
    return generate_latest()


if __name__ == '__main__':
    # Prometheus 메트릭 서버 시작 (포트 8000)
    # start_http_server(8000)
    # print("Prometheus metrics available at http://localhost:8000/metrics")

    # Flask 앱 시작
    print("Flask app starting at http://localhost:5000")
    print("Metrics available at http://localhost:5000/metrics")
    app.run(host='0.0.0.0', port=5000, debug=True)
