# Unisphere_exporter

Unisphere Exporter는 Unity 콘솔에 연결하여 시스템에 관한 `metric`과 `logs`를 수집하여,
Prometheus 또는 Loki와 같은 OpenTelemetry를 지원하는 오픈소스나 상용 백엔드로 수집된 데이터를 전송할 수 있습니다.

## 목적
- ㄴ

## Collector
| Collector       | Default Enabled  | Type     | Description                                           |
|-----------------|------------------|----------|-------------------------------------------------------|
| alert           | false            | `log`    | 스토리지 시스템에서 발생한 Alert 정보 (프로그램 실행으로부터, 한시간 전 데이터부터 수집) |
| basicSystemInfo | true             | `metric` | 스토리지 시스템의 기본 정보                                       |
| capacity        | true             | `metric` | 스토리지 시스템의 용량 정보                                       |
| disk            | false            | `metric` | 스토리지 시스템의 물리 디스크 정보                                   |
| ethernetPort    | false            | `metric` | 스토리지 시스템의 Ethernet Port 정보                            |
| event           | false            | `log`    | 스토리지 시스템에서 발생한 이벤트 정보 (프로그램 실행으로부터, 한시간 전 데이터부터 수집)   |
| fcPort          | false            | `metric` | 스토리지 시스템의 Fibre Channel Port 정보                       |
| host            | false            | `metric` | 스토리지 시스템의 host 정보                                     |
| lun             | true             | `metric` | 스토리지 시스템의 lun 정보                                      |
| metric          | false            | `metric` | 스토리지 시스템의 성능 정보 (RealTimeQuery 이용)                    |



## 언어

- [English](/README.md)
- [Korean](/README_kr.md)
