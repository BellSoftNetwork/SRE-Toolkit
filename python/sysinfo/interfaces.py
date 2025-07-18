"""sysinfo 인터페이스 정의"""

from abc import ABC, abstractmethod
from enum import Enum
from typing import Any, Dict, List, Protocol


class OutputFormat(str, Enum):
    """출력 형식"""

    TABLE = "table"
    JSON = "json"
    YAML = "yaml"


class SystemInfo(Protocol):
    """시스템 정보 프로토콜"""

    os: str
    version: str
    architecture: str
    hostname: str
    python_version: str
    boot_time: str
    uptime: str


class CPUInfo(Protocol):
    """CPU 정보 프로토콜"""

    physical_cores: int
    logical_cores: int
    usage_percent: float
    per_core_usage: List[float]
    frequency_current: str
    frequency_max: str


class MemoryInfo(Protocol):
    """메모리 정보 프로토콜"""

    total: str
    used: str
    available: str
    percent: str
    swap_total: str
    swap_used: str
    swap_free: str
    swap_percent: str


class DiskInfo(Protocol):
    """디스크 정보 프로토콜"""

    device: str
    mountpoint: str
    fstype: str
    total: str
    used: str
    free: str
    percent: str


class NetworkInfo(Protocol):
    """네트워크 정보 프로토콜"""

    interface: str
    status: str
    speed: str
    mtu: int
    addresses: List[Dict[str, str]]


class InfoCollector(ABC):
    """정보 수집기 추상 클래스"""

    @abstractmethod
    def collect_system(self) -> Dict[str, Any]:
        """시스템 정보 수집"""
        pass

    @abstractmethod
    def collect_cpu(self) -> Dict[str, Any]:
        """CPU 정보 수집"""
        pass

    @abstractmethod
    def collect_memory(self) -> Dict[str, Any]:
        """메모리 정보 수집"""
        pass

    @abstractmethod
    def collect_disk(self) -> List[Dict[str, Any]]:
        """디스크 정보 수집"""
        pass

    @abstractmethod
    def collect_network(self) -> Dict[str, Dict[str, Any]]:
        """네트워크 정보 수집"""
        pass


class Formatter(ABC):
    """출력 포맷터 추상 클래스"""

    @abstractmethod
    def format(self, data: Dict[str, Any]) -> Any:
        """데이터를 특정 형식으로 포맷"""
        pass
