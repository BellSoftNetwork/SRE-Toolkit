"""출력 포맷터"""

from rich.table import Table
from typing import Any, Dict, List

from .interfaces import Formatter


class JSONFormatter(Formatter):
    """JSON 형식 포맷터"""

    def format(self, data: Dict[str, Any]) -> Dict[str, Any]:
        """데이터를 JSON 형식으로 포맷"""
        return data


class TableFormatter(Formatter):
    """테이블 형식 포맷터"""

    def format(self, data: Dict[str, Any]) -> Dict[str, Any]:
        """데이터를 Rich 테이블로 포맷"""
        tables = {}

        if "system" in data:
            tables["시스템 정보"] = self._create_info_table(data["system"])

        if "cpu" in data:
            tables["CPU 정보"] = self._create_cpu_table(data["cpu"])

        if "memory" in data:
            tables["메모리 정보"] = self._create_info_table(data["memory"])

        if "disk" in data:
            tables["디스크 정보"] = self._create_disk_table(data["disk"])

        if "network" in data:
            tables["네트워크 정보"] = self._create_network_table(data["network"])

        return tables

    def _create_info_table(self, info: Dict[str, Any]) -> Table:
        """기본 정보 테이블 생성"""
        table = Table(show_header=False)
        table.add_column("항목", style="cyan")
        table.add_column("값", style="green")

        for key, value in info.items():
            if isinstance(value, list):
                value = ", ".join(f"{v:.1f}%" for v in value)
            table.add_row(key, str(value))

        return table

    def _create_cpu_table(self, info: Dict[str, Any]) -> Table:
        """CPU 정보 테이블 생성"""
        table = Table(show_header=False)
        table.add_column("항목", style="cyan")
        table.add_column("값", style="green")

        for key, value in info.items():
            if key == "코어별 사용률":
                formatted_values = []
                for i, usage in enumerate(value):
                    formatted_values.append(f"코어 {i}: {usage:.1f}%")
                value = "\n".join(formatted_values)
            elif isinstance(value, float):
                value = f"{value:.1f}%"
            table.add_row(key, str(value))

        return table

    def _create_disk_table(self, disks: List[Dict[str, Any]]) -> Table:
        """디스크 정보 테이블 생성"""
        table = Table()
        table.add_column("장치", style="cyan")
        table.add_column("마운트 위치", style="yellow")
        table.add_column("파일시스템", style="blue")
        table.add_column("전체 크기", style="green")
        table.add_column("사용 중", style="green")
        table.add_column("여유 공간", style="green")
        table.add_column("사용률", style="magenta")

        for disk in disks:
            table.add_row(
                disk["device"],
                disk["mountpoint"],
                disk["fstype"],
                disk["total"],
                disk["used"],
                disk["free"],
                disk["percent"],
            )

        return table

    def _create_network_table(self, interfaces: Dict[str, Dict[str, Any]]) -> Table:
        """네트워크 정보 테이블 생성"""
        table = Table()
        table.add_column("인터페이스", style="cyan")
        table.add_column("상태", style="yellow")
        table.add_column("속도", style="blue")
        table.add_column("MTU", style="green")
        table.add_column("주소", style="green")

        for iface, data in interfaces.items():
            addresses = []
            for addr in data["addresses"]:
                if addr["family"] == "IPv4":
                    addresses.append(f"IPv4: {addr['address']}")
                elif addr["family"] == "IPv6":
                    addresses.append(f"IPv6: {addr['address'][:30]}...")

            table.add_row(
                iface,
                data["status"],
                data["speed"],
                str(data["mtu"]),
                "\n".join(addresses) if addresses else "주소 없음",
            )

        return table
