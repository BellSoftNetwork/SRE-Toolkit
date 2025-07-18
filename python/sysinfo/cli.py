"""CLI 인터페이스"""

import click
from common import Config, get_logger
from common.logging import LogConfig, LogLevel
from pathlib import Path
from rich.console import Console
from rich.panel import Panel
from rich.table import Table
from typing import Optional

from .collectors import SystemInfoCollector
from .formatters import JSONFormatter, TableFormatter

console = Console()


@click.command()
@click.option("--cpu", is_flag=True, help="CPU 정보 표시")
@click.option("--memory", is_flag=True, help="메모리 정보 표시")
@click.option("--disk", is_flag=True, help="디스크 정보 표시")
@click.option("--network", is_flag=True, help="네트워크 정보 표시")
@click.option("--all", "show_all", is_flag=True, help="모든 정보 표시")
@click.option("--json", "as_json", is_flag=True, help="JSON 형식으로 출력")
@click.option("--config", type=click.Path(exists=True), help="설정 파일 경로")
@click.option("--debug", is_flag=True, help="디버그 모드")
def main(
        cpu: bool,
        memory: bool,
        disk: bool,
        network: bool,
        show_all: bool,
        as_json: bool,
        config: Optional[str],
        debug: bool,
) -> int:
    """시스템 정보를 출력하는 CLI 도구

    옵션 없이 실행하면 기본 시스템 정보만 표시합니다.
    """
    # 로깅 설정
    log_level = LogLevel.DEBUG if debug else LogLevel.ERROR
    logger = get_logger("sysinfo", LogConfig(level=log_level))

    # 설정 로드
    cfg = Config()
    if config:
        try:
            cfg.load_file(Path(config))
            logger.debug(f"설정 파일 로드: {config}")
        except Exception as e:
            logger.error(f"설정 파일 로드 실패: {e}")
            return 1

    try:
        # 정보 수집
        collector = SystemInfoCollector(logger)
        data = collector.collect(
            include_cpu=cpu or show_all,
            include_memory=memory or show_all,
            include_disk=disk or show_all,
            include_network=network or show_all,
        )

        # 출력
        if as_json:
            formatter = JSONFormatter()
            console.print_json(data=formatter.format(data))
        else:
            formatter = TableFormatter()
            output = formatter.format(data)

            for section_name, content in output.items():
                if isinstance(content, Table):
                    console.print(Panel(content, title=section_name, expand=False))
                else:
                    console.print(content)

        return 0

    except Exception as e:
        logger.error(f"오류 발생: {e}", exc_info=True)
        console.print(f"[red]오류 발생: {e}[/red]")
        return 1
