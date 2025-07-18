"""공용 인터페이스 정의"""

from abc import ABC, abstractmethod
from pathlib import Path
from typing import Any, Dict, Optional, Protocol


class ConfigLoader(Protocol):
    """설정 로더 인터페이스"""

    def load(self, path: Path) -> Dict[str, Any]:
        """설정 파일을 로드합니다."""
        ...

    def validate(self, data: Dict[str, Any]) -> bool:
        """설정 데이터를 검증합니다."""
        ...


class Logger(Protocol):
    """로거 인터페이스"""

    def debug(self, message: str, **kwargs: Any) -> None:
        """디버그 메시지를 로깅합니다."""
        ...

    def info(self, message: str, **kwargs: Any) -> None:
        """정보 메시지를 로깅합니다."""
        ...

    def warning(self, message: str, **kwargs: Any) -> None:
        """경고 메시지를 로깅합니다."""
        ...

    def error(self, message: str, **kwargs: Any) -> None:
        """에러 메시지를 로깅합니다."""
        ...

    def critical(self, message: str, **kwargs: Any) -> None:
        """치명적 에러 메시지를 로깅합니다."""
        ...


class ConfigProvider(ABC):
    """설정 제공자 추상 클래스"""

    @abstractmethod
    def get(self, key: str, default: Any = None) -> Any:
        """설정 값을 가져옵니다."""
        pass

    @abstractmethod
    def set(self, key: str, value: Any) -> None:
        """설정 값을 설정합니다."""
        pass

    @abstractmethod
    def exists(self, key: str) -> bool:
        """설정 키가 존재하는지 확인합니다."""
        pass


class MetricsCollector(Protocol):
    """메트릭 수집기 인터페이스"""

    def increment(self, metric: str, value: float = 1.0, tags: Optional[Dict[str, str]] = None) -> None:
        """카운터를 증가시킵니다."""
        ...

    def gauge(self, metric: str, value: float, tags: Optional[Dict[str, str]] = None) -> None:
        """게이지 값을 설정합니다."""
        ...

    def histogram(self, metric: str, value: float, tags: Optional[Dict[str, str]] = None) -> None:
        """히스토그램 값을 기록합니다."""
        ...
