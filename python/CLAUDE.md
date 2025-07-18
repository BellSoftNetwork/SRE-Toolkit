# Python 유틸리티 개발 가이드
## 🎯 개요
이 문서는 Python 기반 SRE 유틸리티를 개발하는 개발자를 위한 가이드입니다. 사용자 가이드는 [README.md](./README.md)를 참조하세요.



## 🏗️ 아키텍처
### 플랫 레이아웃 구조
```
python/
├── common/                    # 공용 모듈
│   ├── __init__.py
│   ├── interfaces.py          # 인터페이스 정의
│   ├── logging.py             # 로깅 모듈
│   ├── config.py              # 설정 관리
│   ├── exceptions.py          # 공용 예외
│   └── utils.py               # 유틸리티 함수
├── sysinfo/                   # 시스템 정보 도구
│   ├── __init__.py
│   ├── __main__.py            # 진입점
│   ├── cli.py                 # CLI 인터페이스
│   ├── interfaces.py          # 도구별 인터페이스
│   ├── collectors.py          # 데이터 수집 로직
│   ├── formatters.py          # 출력 포맷터
│   └── README.md              # 도구별 사용자 문서
├── tests/                     # 테스트 코드
│   ├── unit/                  # 단위 테스트
│   └── integration/           # 통합 테스트
├── pyproject.toml             # 프로젝트 설정
├── .python-version            # Python 버전 (3.12)
├── README.md                  # 사용자 가이드
└── CLAUDE.md                  # 개발자 가이드 (이 문서)
```

### 설계 원칙
1. **도구 독립성**: 각 도구는 독립적으로 실행 가능
2. **공통 모듈 재사용**: 로깅, 설정, 유틸리티는 공용 모듈 사용
3. **인터페이스 기반 설계**: Protocol과 ABC로 확장 가능한 구조
4. **의존성 주입**: 느슨한 결합으로 테스트와 확장 용이



## 🚀 개발 환경 설정
### 1. 사전 요구사항
```bash
# uv 설치 (최초 1회)
curl -LsSf https://astral.sh/uv/install.sh | sh
```

### 2. 개발 환경 구성
```bash
# Python 디렉토리로 이동
cd python

# 가상환경 생성
uv venv

# 가상환경 활성화
source .venv/bin/activate  # Windows: .venv\Scripts\activate

# 개발 모드로 패키지 설치
uv pip install -e .

# 개발 의존성 설치
uv pip install -e ".[dev]"
```



## 📦 공용 모듈 API
### 로깅 모듈 (`common.logging`)
```python
from common import get_logger
from common.logging import LogConfig, LogLevel

# 기본 로거 생성
logger = get_logger("my_tool")

# 커스텀 설정으로 로거 생성
config = LogConfig(
    level=LogLevel.DEBUG,
    use_rich=True,  # Rich 라이브러리 사용
    file_path=Path("app.log")  # 파일 로깅
)
logger = get_logger("my_tool", config)

# 사용 예시
logger.debug("디버그 메시지")
logger.info("정보 메시지")
logger.warning("경고 메시지")
logger.error("오류 메시지", error_code=500)
logger.critical("심각한 오류")
```

### 설정 관리 (`common.config`)
```python
from common import Config

# 설정 인스턴스 생성
config = Config()

# YAML/JSON 파일에서 설정 로드
config.load_file("config.yaml")

# 환경변수에서 설정 로드 (prefix 지정)
config.load_env("MYTOOL_")  # MYTOOL_DATABASE_HOST → database.host

# 설정 값 읽기
db_host = config.get("database.host", default="localhost")
db_port = config.get("database.port", default=5432)
api_timeout = config.get("api.timeout", default=30)

# 중첩된 설정 읽기
redis_config = config.get("cache.redis", default={})

# 설정 값 쓰기
config.set("api.timeout", 60)
config.set("database.pool_size", 10)
```

### 유틸리티 함수 (`common.utils`)
```python
from common.utils import (
    retry, format_bytes, format_duration,
    get_timestamp, ensure_directory, run_command
)

# 재시도 데코레이터
@retry(max_attempts=3, delay=1.0, backoff=2.0)
def unreliable_api_call():
    response = requests.get("https://api.example.com")
    return response.json()

# 바이트 포맷팅
print(format_bytes(1048576))      # "1.00 MB"
print(format_bytes(1073741824))   # "1.00 GB"

# 시간 포맷팅
print(format_duration(3661))      # "1시간 1분 1초"
print(format_duration(90))        # "1분 30초"

# 타임스탬프 생성
print(get_timestamp())            # "2025-01-18T10:30:45"
print(get_timestamp(utc=True))    # UTC 시간

# 디렉토리 생성 (재귀적)
ensure_directory("logs/2025/01")

# 외부 명령 실행
result = run_command(["ls", "-la"], timeout=30)
if result.success:
    print(result.stdout)
else:
    print(f"Error: {result.stderr}")
```

### 예외 클래스 (`common.exceptions`)
```python
from common.exceptions import (
    SREToolkitError, ConfigError, ValidationError,
    NetworkError, TimeoutError
)

# 커스텀 예외 발생
try:
    if not config_file.exists():
        raise ConfigError(f"설정 파일을 찾을 수 없습니다: {config_file}")
    
    if port < 1 or port > 65535:
        raise ValidationError(f"잘못된 포트 번호: {port}")
        
except SREToolkitError as e:
    logger.error(f"도구 오류: {e}")
```



## 🛠️ 새 도구 개발 가이드
### 1. 도구 구조 생성
```bash
# python 디렉토리에서 실행
TOOL_NAME="mytool"  # 도구 이름 설정

# 디렉토리 생성
mkdir -p ${TOOL_NAME}

# 기본 파일 생성
touch ${TOOL_NAME}/{__init__.py,__main__.py,cli.py,interfaces.py,README.md}
```

### 2. 진입점 구현 (`__main__.py`)
```python
"""도구 진입점"""
import sys
from .cli import main

if __name__ == "__main__":
    sys.exit(main())
```

### 3. CLI 인터페이스 구현 (`cli.py`)
```python
"""CLI 인터페이스"""
import click
from typing import Optional
from pathlib import Path

from common import get_logger, Config
from common.logging import LogConfig, LogLevel

logger = get_logger(__name__)

@click.command()
@click.option("--config", "-c", type=click.Path(exists=True), 
              help="설정 파일 경로")
@click.option("--output", "-o", type=click.Choice(["text", "json", "yaml"]), 
              default="text", help="출력 형식")
@click.option("--debug", is_flag=True, help="디버그 모드")
@click.version_option(version="0.1.0")
def main(config: Optional[str], output: str, debug: bool) -> int:
    """도구 설명"""
    # 로깅 설정
    log_level = LogLevel.DEBUG if debug else LogLevel.INFO
    log_config = LogConfig(level=log_level, use_rich=True)
    logger = get_logger("mytool", log_config)
    
    try:
        # 설정 로드
        app_config = Config()
        if config:
            app_config.load_file(config)
        app_config.load_env("MYTOOL_")
        
        # 도구 로직 실행
        logger.info("도구 실행 시작")
        # ... 실제 로직 구현 ...
        
        return 0
        
    except Exception as e:
        logger.error(f"실행 중 오류 발생: {e}")
        return 1
```

### 4. 인터페이스 정의 (`interfaces.py`)
```python
"""도구별 인터페이스 정의"""
from typing import Protocol, Dict, Any
from abc import ABC, abstractmethod

class DataCollector(Protocol):
    """데이터 수집 인터페이스"""
    def collect(self) -> Dict[str, Any]:
        """데이터를 수집하여 반환"""
        ...

class OutputFormatter(ABC):
    """출력 포맷터 추상 클래스"""
    @abstractmethod
    def format(self, data: Dict[str, Any]) -> str:
        """데이터를 문자열로 포맷"""
        pass
```

### 5. pyproject.toml에 진입점 추가
```toml
[project.scripts]
mytool = "mytool:main"
```

이 설정으로 `uv pip install -e .` 실행 시 `mytool` 명령어가 자동으로 생성되어,
가상환경에서 바로 `mytool` 명령어로 실행할 수 있습니다.

### 6. 도구 문서 작성 (`README.md`)
```markdown
# MyTool

도구에 대한 간단한 설명

## 사용법

\`\`\`bash
# 기본 실행
mytool

# 옵션과 함께 실행
mytool --output json --debug
\`\`\`

## 옵션

- `--config, -c`: 설정 파일 경로
- `--output, -o`: 출력 형식 (text, json, yaml)
- `--debug`: 디버그 모드 활성화
- `--help`: 도움말 표시

## 출력 예시

\`\`\`
[출력 예시 내용]
\`\`\`
```



## 🧪 테스트 작성
### 단위 테스트 예시
```python
# tests/unit/test_mytool.py
import pytest
from unittest.mock import Mock, patch

from mytool.collectors import SystemCollector

class TestSystemCollector:
    """SystemCollector 단위 테스트"""
    
    def test_collect_returns_dict(self):
        """collect 메서드가 딕셔너리를 반환하는지 테스트"""
        collector = SystemCollector()
        result = collector.collect()
        
        assert isinstance(result, dict)
        assert "system" in result
        
    @patch("psutil.cpu_percent")
    def test_collect_cpu_info(self, mock_cpu):
        """CPU 정보 수집 테스트"""
        mock_cpu.return_value = 50.0
        
        collector = SystemCollector()
        result = collector.collect()
        
        assert result["cpu"]["usage"] == 50.0
        mock_cpu.assert_called_once()
```

### 통합 테스트 예시
```python
# tests/integration/test_mytool_cli.py
import pytest
from click.testing import CliRunner

from mytool.cli import main

class TestMyToolCLI:
    """MyTool CLI 통합 테스트"""
    
    def test_cli_basic_run(self):
        """기본 실행 테스트"""
        runner = CliRunner()
        result = runner.invoke(main)
        
        assert result.exit_code == 0
        assert "도구 실행 완료" in result.output
        
    def test_cli_with_json_output(self):
        """JSON 출력 테스트"""
        runner = CliRunner()
        result = runner.invoke(main, ["--output", "json"])
        
        assert result.exit_code == 0
        # JSON 파싱 가능한지 확인
        import json
        json.loads(result.output)
```



## 🔍 코드 품질 관리
### 코드 포맷팅
```bash
# Black으로 코드 포맷팅
black mytool/

# 변경사항 미리보기
black --diff src/sre_toolkit/tools/mytool/
```

### 린팅
```bash
# Ruff로 코드 검사
ruff check mytool/

# 자동 수정 가능한 문제 수정
ruff check --fix src/sre_toolkit/tools/mytool/
```

### 타입 체크
```bash
# mypy로 타입 검사
mypy mytool/

# 엄격한 모드로 검사
mypy --strict src/sre_toolkit/tools/mytool/
```

### 테스트 실행
```bash
# 특정 도구의 테스트만 실행
pytest tests/unit/test_mytool.py -v

# 커버리지와 함께 실행
pytest tests/ -v --cov=mytool --cov-report=term-missing
```



## 📝 문서화 가이드
### 도구 README 구조
1. **개요**: 도구의 목적과 주요 기능
2. **사용법**: 실행 명령 예시 (복사해서 바로 실행 가능하게)
3. **옵션**: 모든 CLI 옵션 설명
4. **출력 예시**: 실제 출력 결과
5. **설정**: 설정 파일이나 환경변수 사용법
6. **문제 해결**: 자주 발생하는 문제와 해결방법

### Docstring 작성
```python
def process_data(data: Dict[str, Any], format: str = "json") -> str:
    """데이터를 지정된 형식으로 처리합니다.
    
    Args:
        data: 처리할 데이터 딕셔너리
        format: 출력 형식 ("json", "yaml", "text")
        
    Returns:
        포맷팅된 문자열
        
    Raises:
        ValueError: 지원하지 않는 형식을 지정한 경우
        
    Example:
        >>> data = {"name": "test", "value": 123}
        >>> result = process_data(data, format="json")
        >>> print(result)
        {"name": "test", "value": 123}
    """
```



## 🚀 배포 체크리스트
### 도구 배포 전 확인사항
- [ ] 모든 테스트 통과
- [ ] 코드 포맷팅 완료 (black)
- [ ] 린팅 통과 (ruff)
- [ ] 타입 체크 통과 (mypy)
- [ ] README.md 작성 완료
- [ ] 버전 정보 업데이트
- [ ] CHANGELOG 업데이트 (있는 경우)

### 빌드 및 배포
```bash
# 패키지 빌드
uv build

# 생성된 파일 확인
ls dist/

# 로컬 설치 테스트
uv pip install dist/*.whl

# 설치 확인
mytool --version
```
