# LogQL 쿼리 예시 모음

이 문서는 Loki에서 사용하는 LogQL 쿼리 예시를 정리한 것입니다.

## 목차
- [기본 쿼리](#기본-쿼리)
- [필터링](#필터링)
- [파싱](#파싱)
- [집계](#집계)
- [메트릭 쿼리](#메트릭-쿼리)
- [실전 예시](#실전-예시)

---

## 기본 쿼리

### 로그 스트림 선택

```logql
# 특정 앱의 로그
{app="backend"}

# 여러 레이블로 필터링
{app="backend", env="production"}

# 레이블 정규표현식 매칭
{app=~"backend|frontend"}

# 레이블 제외
{app!="test"}
```

### 시간 범위

```logql
# 최근 5분
{app="backend"} [5m]

# 최근 1시간
{app="backend"} [1h]

# 특정 시간 범위는 Grafana UI에서 설정
```

---

## 필터링

### 텍스트 필터

```logql
# 특정 텍스트 포함
{app="backend"} |= "error"

# 특정 텍스트 제외
{app="backend"} != "debug"

# 대소문자 구분 없이 검색
{app="backend"} |~ "(?i)error"

# 여러 조건 (AND)
{app="backend"} |= "error" |= "database"

# 여러 조건 (OR - 정규표현식 사용)
{app="backend"} |~ "error|fail|exception"
```

### 정규표현식 필터

```logql
# 정규표현식 매칭
{app="backend"} |~ "error.*database"

# 정규표현식 제외
{app="backend"} !~ "debug|trace"

# IP 주소 찾기
{app="nginx"} |~ "\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}\\.\\d{1,3}"

# HTTP 상태 코드 5xx
{app="nginx"} |~ "\" 5\\d{2} "
```

---

## 파싱

### JSON 파싱

```logql
# JSON 필드 파싱
{app="backend"} | json

# 특정 JSON 필드만 추출
{app="backend"} | json level, message, user_id

# 중첩된 JSON 필드
{app="backend"} | json response_body="response.body"

# JSON 파싱 후 필터링
{app="backend"} | json | level="error"

# JSON 필드 값으로 필터링
{app="backend"} | json | response_time > 1000
```

### Logfmt 파싱

```logql
# logfmt 파싱 (key=value 형식)
{app="backend"} | logfmt

# 특정 필드만 추출
{app="backend"} | logfmt level, msg, user_id

# 파싱 후 필터링
{app="backend"} | logfmt | level="error"
```

### 패턴 파싱

```logql
# 패턴 매칭으로 필드 추출
{app="nginx"} | pattern "<ip> - - <_> \"<method> <uri> <_>\" <status> <size>"

# 추출된 필드로 필터링
{app="nginx"} | pattern "<ip> - - <_> \"<method> <uri> <_>\" <status> <size>" | status >= 400
```

### 정규표현식 파싱

```logql
# 정규표현식으로 필드 추출
{app="backend"} | regexp "user_id=(?P<user_id>\\d+)"

# 여러 필드 추출
{app="backend"} | regexp "level=(?P<level>\\w+).*msg=\"(?P<message>[^\"]+)\""
```

---

## 집계

### Line Format

```logql
# 로그 라인 재포맷
{app="backend"} | json | line_format "{{.level}}: {{.message}}"

# 조건부 포맷
{app="backend"} | json | line_format "{{ if eq .level \"error\" }}🔴{{ else }}{{ end }} {{.message}}"
```

### Label Format

```logql
# 새로운 레이블 추가
{app="backend"} | json | label_format level="{{.level}}"

# 레이블 값 변환
{app="backend"} | json | label_format status="{{ if eq .status_code \"200\" }}success{{ else }}failure{{ end }}"
```

---

## 메트릭 쿼리

### 카운트

```logql
# 로그 라인 수
count_over_time({app="backend"}[5m])

# 에러 로그 수
count_over_time({app="backend"} |= "error" [5m])

# JSON 필터링 후 카운트
count_over_time({app="backend"} | json | level="error" [5m])

# 여러 스트림 합계
sum(count_over_time({app="backend"}[5m]))
```

### Rate

```logql
# 초당 로그 발생률
rate({app="backend"}[5m])

# 에러 발생률
rate({app="backend"} |= "error" [5m])

# 에러율 (%)
sum(rate({app="backend"} | json | level="error" [5m]))
/
sum(rate({app="backend"} [5m])) * 100
```

### 집계 함수

```logql
# 평균
avg_over_time({app="backend"} | json | unwrap response_time [5m])

# 최대값
max_over_time({app="backend"} | json | unwrap response_time [5m])

# 최소값
min_over_time({app="backend"} | json | unwrap response_time [5m])

# 합계
sum_over_time({app="backend"} | json | unwrap request_size [5m])

# Quantile (p95)
quantile_over_time(0.95, {app="backend"} | json | unwrap response_time [5m])
```

### Bytes 메트릭

```logql
# 로그 크기 (bytes)
bytes_over_time({app="backend"}[5m])

# 초당 로그 크기 (bytes/s)
bytes_rate({app="backend"}[5m])
```

---

## 실전 예시

### 에러 분석

```logql
# 에러 로그 실시간 조회
{app="backend"} |= "error"

# 에러 레벨별 분포
sum by (level) (count_over_time({app="backend"} | json | level=~"error|fatal" [5m]))

# 특정 에러 메시지 검색
{app="backend"} |~ "(?i)(exception|error|fail)" | json | message =~ "database"

# 가장 많이 발생한 에러 Top 5
topk(5, sum by (error_type) (count_over_time({app="backend"} | json | level="error" [1h])))
```

### 성능 분석

```logql
# 응답 시간 p95
quantile_over_time(0.95, {app="api"} | json | unwrap response_time [5m])

# 느린 요청 찾기 (1초 이상)
{app="api"} | json | response_time > 1000

# 엔드포인트별 평균 응답 시간
avg by (endpoint) (avg_over_time({app="api"} | json | unwrap response_time [5m]))

# 응답 시간 히스토그램
histogram_over_time({app="api"} | json | unwrap response_time [5m])
```

### HTTP 액세스 로그 분석

```logql
# 404 에러 찾기
{app="nginx"} |~ "\" 404 "

# 5xx 에러 발생률
sum(rate({app="nginx"} |~ "\" 5\\d{2} " [5m]))

# 상태 코드별 분포
sum by (status) (count_over_time(
  {app="nginx"} | pattern "<_> \"<_>\" <status> <_>" [5m]
))

# IP별 요청 수
sum by (ip) (count_over_time(
  {app="nginx"} | pattern "<ip> <_>" [1h]
))

# 엔드포인트별 트래픽
sum by (uri) (count_over_time(
  {app="nginx"} | pattern "<_> \"<method> <uri> <_>\"" [5m]
))
```

### 보안 로그 분석

```logql
# 인증 실패 로그
{app="auth"} |= "authentication failed"

# 5분 내 5회 이상 로그인 실패한 IP
sum by (ip) (count_over_time(
  {app="auth"} | json | message="login failed" [5m]
)) > 5

# SQL 인젝션 시도 감지
{app="api"} |~ "(?i)(union|select|drop|delete|insert).*from"

# 비정상적인 API 호출 (분당 100회 이상)
sum by (ip) (rate({app="api"}[1m])) > 100

# SSH 로그인 실패
{job="syslog"} |= "sshd" |= "Failed password"
```

### 사용자 행동 추적

```logql
# 특정 사용자의 모든 액션
{app="backend"} | json | user_id="12345"

# 사용자별 요청 수
sum by (user_id) (count_over_time({app="backend"} | json [1h]))

# 주문 완료 이벤트
{app="backend"} | json | event="order_completed"

# 사용자 여정 추적
{app="backend"} | json | session_id="abc123" | line_format "{{.timestamp}} {{.event}} {{.page}}"
```

### 비즈니스 메트릭

```logql
# 주문 완료 수 (시간당)
sum(count_over_time({app="backend"} | json | event="order_completed" [1h]))

# 결제 실패율
sum(count_over_time({app="payment"} | json | status="failed" [5m]))
/
sum(count_over_time({app="payment"} | json [5m])) * 100

# 신규 가입자 수
sum(count_over_time({app="backend"} | json | event="user_registered" [1h]))

# 장바구니 이탈
sum(count_over_time({app="backend"} | json | event="cart_abandoned" [1h]))
```

### 애플리케이션 모니터링

```logql
# 특정 예외 발생 수
count_over_time({app="backend"} |~ "NullPointerException" [5m])

# 데이터베이스 쿼리 에러
{app="backend"} |= "database" |= "error"

# 외부 API 호출 실패
{app="backend"} | json | external_api_status != "200"

# 메모리 부족 경고
{app="backend"} |~ "(?i)(out of memory|oom)"

# 타임아웃 에러
{app="backend"} |~ "(?i)(timeout|timed out)"
```

### 로그 비교 (대시보드용)

```logql
# 현재 vs 1일 전 에러율 비교
# Panel 1: 현재
sum(rate({app="backend"} | json | level="error" [5m]))

# Panel 2: 1일 전 (Grafana에서 time shift 사용)
sum(rate({app="backend"} | json | level="error" [5m]))
```

---

## 고급 쿼리

### Unwrap

```logql
# 숫자 필드 값을 메트릭으로 사용
{app="api"} | json | unwrap response_time

# Unwrap + 집계
avg_over_time({app="api"} | json | unwrap response_time [5m])

# 조건부 Unwrap
{app="api"} | json | endpoint="/api/users" | unwrap response_time
```

### 복잡한 필터링

```logql
# 다중 조건
{app="backend"}
| json
| level="error"
| user_id != ""
| response_time > 1000

# 정규표현식 + JSON
{app="api"}
| json
| endpoint =~ "/api/(users|orders)"
| status_code >= 400
```

### Label Operations

```logql
# 레이블 추가 및 변환
{app="backend"}
| json
| label_format
    severity="{{ if eq .level \"error\" }}high{{ else }}low{{ end }}",
    user="{{ .user_id }}"

# 레이블로 그룹화
sum by (severity) (
  rate({app="backend"}
  | json
  | label_format severity="{{ .level }}" [5m])
)
```

---

## 성능 최적화 팁

### 쿼리 최적화

```logql
# 나쁜 예: 너무 넓은 범위
{app="backend"}

# 좋은 예: 레이블로 좁히기
{app="backend", env="production", namespace="default"}

# 나쁜 예: 파싱 후 필터
{app="backend"} | json | level="error"

# 좋은 예: 텍스트 필터 먼저
{app="backend"} |= "error" | json | level="error"
```

### 시간 범위

```logql
# 짧은 시간 범위 사용 (5분 권장)
[5m]

# 긴 시간 범위는 피하기 (메모리 부족 가능)
[24h]  # 주의!
```

---

## 참고 자료

- [LogQL Documentation](https://grafana.com/docs/loki/latest/logql/)
- [LogQL Cheat Sheet](https://megamorf.gitlab.io/cheat-sheets/loki/)
- [LogQL Examples](https://grafana.com/docs/loki/latest/logql/examples/)
