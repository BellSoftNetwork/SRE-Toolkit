"""시스템 정보 수집기"""

import platform
import psutil
import socket
from common.interfaces import Logger
from common.utils import format_bytes, format_duration
from datetime import datetime, timezone
from typing import Any, Dict, List

from .interfaces import InfoCollector


class SystemInfoCollector(InfoCollector):
    """시스템 정보 수집기 구현"""

    def __init__(self, logger: Logger) -> None:
        self.logger = logger

    def collect(
            self,
            include_cpu: bool = False,
            include_memory: bool = False,
            include_disk: bool = False,
            include_network: bool = False,
    ) -> Dict[str, Any]:
        """전체 시스템 정보 수집

        Args:
            include_cpu: CPU 정보 포함 여부
            include_memory: 메모리 정보 포함 여부
            include_disk: 디스크 정보 포함 여부
            include_network: 네트워크 정보 포함 여부

        Returns:
            수집된 시스템 정보
        """
        data = {"system": self.collect_system()}

        if include_cpu:
            data["cpu"] = self.collect_cpu()
        if include_memory:
            data["memory"] = self.collect_memory()
        if include_disk:
            data["disk"] = self.collect_disk()
        if include_network:
            data["network"] = self.collect_network()

        return data

    def collect_system(self) -> Dict[str, Any]:
        """기본 시스템 정보 수집"""
        self.logger.debug("시스템 정보 수집 시작")

        boot_time = datetime.fromtimestamp(psutil.boot_time(), tz=timezone.utc)
        current_time = datetime.now(timezone.utc)
        uptime_seconds = (current_time - boot_time).total_seconds()

        return {
            "운영체제": platform.system(),
            "OS 버전": platform.version(),
            "아키텍처": platform.machine(),
            "호스트명": socket.gethostname(),
            "Python 버전": platform.python_version(),
            "부팅 시간": boot_time.strftime("%Y-%m-%d %H:%M:%S %Z"),
            "가동 시간": format_duration(uptime_seconds),
        }

    def collect_cpu(self) -> Dict[str, Any]:
        """CPU 정보 수집"""
        self.logger.debug("CPU 정보 수집 시작")

        cpu_percent = psutil.cpu_percent(interval=1, percpu=True)
        cpu_freq = psutil.cpu_freq()

        return {
            "물리 코어": psutil.cpu_count(logical=False),
            "논리 코어": psutil.cpu_count(logical=True),
            "전체 사용률": psutil.cpu_percent(interval=1),
            "코어별 사용률": cpu_percent,
            "현재 주파수": f"{cpu_freq.current:.2f} MHz" if cpu_freq else "정보 없음",
            "최대 주파수": f"{cpu_freq.max:.2f} MHz" if cpu_freq and cpu_freq.max > 0 else "정보 없음",
        }

    def collect_memory(self) -> Dict[str, Any]:
        """메모리 정보 수집"""
        self.logger.debug("메모리 정보 수집 시작")

        memory = psutil.virtual_memory()
        swap = psutil.swap_memory()

        return {
            "전체 메모리": format_bytes(memory.total),
            "사용 중": format_bytes(memory.used),
            "여유 메모리": format_bytes(memory.available),
            "사용률": f"{memory.percent}%",
            "스왑 전체": format_bytes(swap.total),
            "스왑 사용": format_bytes(swap.used),
            "스왑 여유": format_bytes(swap.free),
            "스왑 사용률": f"{swap.percent}%",
        }

    def collect_disk(self) -> List[Dict[str, Any]]:
        """디스크 정보 수집"""
        self.logger.debug("디스크 정보 수집 시작")

        disks = []
        partitions = psutil.disk_partitions()

        for partition in partitions:
            try:
                usage = psutil.disk_usage(partition.mountpoint)
                disks.append(
                    {
                        "device": partition.device,
                        "mountpoint": partition.mountpoint,
                        "fstype": partition.fstype,
                        "total": format_bytes(usage.total),
                        "used": format_bytes(usage.used),
                        "free": format_bytes(usage.free),
                        "percent": f"{usage.percent}%",
                    }
                )
            except PermissionError:
                self.logger.warning(f"권한 없음: {partition.mountpoint}")
                continue

        return disks

    def collect_network(self) -> Dict[str, Dict[str, Any]]:
        """네트워크 정보 수집"""
        self.logger.debug("네트워크 정보 수집 시작")

        interfaces = {}
        stats = psutil.net_if_stats()
        addrs = psutil.net_if_addrs()

        for iface, stat in stats.items():
            addresses = []
            if iface in addrs:
                for addr in addrs[iface]:
                    if addr.family == socket.AF_INET:
                        addresses.append(
                            {
                                "family": "IPv4",
                                "address": addr.address,
                                "netmask": addr.netmask,
                            }
                        )
                    elif addr.family == socket.AF_INET6:
                        addresses.append(
                            {
                                "family": "IPv6",
                                "address": addr.address,
                            }
                        )

            interfaces[iface] = {
                "status": "활성" if stat.isup else "비활성",
                "speed": f"{stat.speed} Mbps" if stat.speed > 0 else "알 수 없음",
                "mtu": stat.mtu,
                "addresses": addresses,
            }

        return interfaces
