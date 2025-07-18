# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with sysinfo.

## 프로젝트 개요
sysinfo는 시스템의 하드웨어 및 소프트웨어 정보를 수집하여 표시하는 Python 기반 CLI 도구입니다.

## 개발 환경 설정

### 사전 요구사항
- Python 3.12 이상
- uv (Python 패키지 관리자)
- psutil 라이브러리에 대한 시스템 권한

### 프로젝트 초기화
```bash
# 저장소 클론
git clone https://gitlab.bellsoft.net/devops/sre-workbench.git
cd sre-workbench/python

# 가상환경 생성 및 활성화
uv venv
source .venv/bin/activate  # Windows: .venv\Scripts\activate

# 개발 모드로 설치
uv pip install -e ".[dev]"
```

## 프로젝트 구조

### 디렉토리 구조
```
sysinfo/
├── __init__.py          # 패키지 초기화
├── __main__.py          # 진입점 (python -m sysinfo)
├── cli.py               # CLI 인터페이스 및 메인 로직
├── interfaces.py        # 인터페이스 정의
├── collectors.py        # 시스템 정보 수집 모듈
├── formatters.py        # 출력 포맷터
├── config.py            # 설정 관리
├── exceptions.py        # 커스텀 예외
├── CLAUDE.md            # 개발자 가이드 (이 파일)
└── README.md            # 사용자 가이드
```

### 핵심 컴포넌트

#### Collectors
시스템 정보를 수집하는 모듈들입니다.
```python
# collectors.py 구조
class SystemCollector:
    """기본 시스템 정보 수집"""
    def collect() -> Dict[str, Any]

class CPUCollector:
    """CPU 정보 수집"""
    def collect() -> Dict[str, Any]

class MemoryCollector:
    """메모리 정보 수집"""
    def collect() -> Dict[str, Any]

class DiskCollector:
    """디스크 정보 수집"""
    def collect() -> Dict[str, Any]

class NetworkCollector:
    """네트워크 정보 수집"""
    def collect() -> Dict[str, Any]
```

#### Formatters
수집된 정보를 다양한 형식으로 출력합니다.
```python
# formatters.py 구조
class OutputFormatter(ABC):
    """출력 포맷터 추상 클래스"""
    @abstractmethod
    def format(self, data: Dict[str, Any]) -> str

class TextFormatter(OutputFormatter):
    """Rich 라이브러리를 사용한 텍스트 출력"""
    
class JSONFormatter(OutputFormatter):
    """JSON 형식 출력"""
```

## 코드 작성 가이드

### 새로운 Collector 추가
1. `collectors.py`에 새 클래스 추가
2. `DataCollector` 프로토콜 구현
3. 에러 처리 포함
4. 테스트 작성

예시:
```python
class ProcessCollector:
    """프로세스 정보 수집"""
    
    def collect(self) -> Dict[str, Any]:
        try:
            processes = []
            for proc in psutil.process_iter(['pid', 'name', 'cpu_percent']):
                processes.append({
                    'pid': proc.info['pid'],
                    'name': proc.info['name'],
                    'cpu': proc.info['cpu_percent']
                })
            
            return {
                'total': len(processes),
                'top_5': sorted(processes, 
                              key=lambda x: x['cpu'], 
                              reverse=True)[:5]
            }
        except Exception as e:
            logger.warning(f"프로세스 정보 수집 실패: {e}")
            return {'error': str(e)}
```

### 새로운 출력 형식 추가
1. `formatters.py`에 새 Formatter 클래스 추가
2. `OutputFormatter` 추상 클래스 상속
3. `cli.py`의 출력 옵션에 추가

예시:
```python
class YAMLFormatter(OutputFormatter):
    """YAML 형식 출력"""
    
    def format(self, data: Dict[str, Any]) -> str:
        import yaml
        return yaml.dump(data, 
                       allow_unicode=True, 
                       default_flow_style=False)
```

## 테스트

### 단위 테스트 실행
```bash
# 특정 모듈 테스트
pytest tests/unit/test_sysinfo_collectors.py -v

# 커버리지 확인
pytest tests/unit/test_sysinfo*.py --cov=sysinfo --cov-report=term-missing
```

### 통합 테스트
```bash
# CLI 테스트
pytest tests/integration/test_sysinfo_cli.py -v
```

### 테스트 작성 예시
```python
# tests/unit/test_sysinfo_collectors.py
import pytest
from unittest.mock import patch, MagicMock

from sysinfo.collectors import CPUCollector

class TestCPUCollector:
    @patch('psutil.cpu_count')
    @patch('psutil.cpu_freq')
    @patch('psutil.cpu_percent')
    def test_collect_cpu_info(self, mock_percent, mock_freq, mock_count):
        # Mock 설정
        mock_count.side_effect = [4, 8]  # 물리, 논리 코어
        mock_freq.return_value = MagicMock(current=2600.0)
        mock_percent.return_value = 25.5
        
        # 테스트 실행
        collector = CPUCollector()
        result = collector.collect()
        
        # 검증
        assert result['물리 코어'] == 4
        assert result['논리 코어'] == 8
        assert result['사용률'] == 25.5
```

## 디버깅

### 로그 레벨 설정
```bash
# 디버그 모드로 실행
sysinfo --debug --all

# 환경변수로 설정
export SYSINFO_LOG_LEVEL=DEBUG
sysinfo --all
```

### 특정 Collector 디버깅
```python
# cli.py에서 개별 collector 테스트
if __name__ == "__main__":
    from collectors import CPUCollector
    collector = CPUCollector()
    print(collector.collect())
```

### 성능 프로파일링
```bash
# cProfile 사용
python -m cProfile -s cumtime -m sysinfo --all
```

## 의존성 관리

### 핵심 의존성
- `psutil`: 시스템 정보 수집
- `click`: CLI 인터페이스
- `rich`: 터미널 출력 포맷팅

### 개발 의존성
- `pytest`: 테스트 프레임워크
- `pytest-cov`: 커버리지 측정
- `black`: 코드 포맷터
- `ruff`: 린터
- `mypy`: 타입 체커

### 의존성 업데이트
```bash
# 의존성 확인
uv pip list --outdated

# 특정 패키지 업데이트
uv pip install --upgrade psutil
```

## 릴리스 프로세스

### 버전 업데이트
1. `pyproject.toml`의 버전 수정
2. `cli.py`의 `@click.version_option` 업데이트
3. CHANGELOG 업데이트 (있는 경우)

### 빌드 및 테스트
```bash
# 전체 테스트 실행
pytest tests/ -v

# 빌드
uv build

# 로컬 설치 테스트
uv pip install dist/*.whl
```

## 주의사항

### 권한 문제
- 일부 시스템 정보는 관리자 권한이 필요할 수 있음
- Windows에서는 WMI 접근 권한 필요
- Linux에서는 /proc 파일시스템 읽기 권한 필요

### 플랫폼 호환성
- Windows, Linux, macOS 지원
- 플랫폼별 차이점 처리 필요
- psutil이 지원하지 않는 정보는 graceful하게 처리

### 성능 고려사항
- 네트워크 정보 수집은 시간이 오래 걸릴 수 있음
- 대용량 디스크 스캔 시 타임아웃 고려
- 병렬 처리로 성능 개선 가능

## 트러블슈팅

### psutil 설치 실패
```bash
# 시스템 패키지 설치 (Ubuntu/Debian)
sudo apt-get install python3-dev

# 재설치
uv pip install --force-reinstall psutil
```

### ImportError 발생
```bash
# 개발 모드 재설치
cd python
uv pip install -e .
```

### 권한 오류
```bash
# Linux/macOS
sudo sysinfo --all

# Windows (관리자 권한으로 실행)
# PowerShell을 관리자 권한으로 열고 실행
```
