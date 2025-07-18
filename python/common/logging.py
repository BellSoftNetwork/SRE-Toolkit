"""로깅 공용 모듈"""

import sys
from enum import Enum
from pathlib import Path
from rich.console import Console
from rich.logging import RichHandler
from typing import Any, Dict, Optional, Union

import logging
from .interfaces import Logger


class LogLevel(str, Enum):
    """로그 레벨"""

    DEBUG = "DEBUG"
    INFO = "INFO"
    WARNING = "WARNING"
    ERROR = "ERROR"
    CRITICAL = "CRITICAL"


class LoggerImpl:
    """Logger 인터페이스 구현체"""

    def __init__(self, logger: logging.Logger) -> None:
        self._logger = logger

    def debug(self, message: str, **kwargs: Any) -> None:
        self._logger.debug(message, extra=kwargs)

    def info(self, message: str, **kwargs: Any) -> None:
        self._logger.info(message, extra=kwargs)

    def warning(self, message: str, **kwargs: Any) -> None:
        self._logger.warning(message, extra=kwargs)

    def error(self, message: str, **kwargs: Any) -> None:
        self._logger.error(message, extra=kwargs)

    def critical(self, message: str, **kwargs: Any) -> None:
        self._logger.critical(message, extra=kwargs)


class LogConfig:
    """로깅 설정"""

    def __init__(
            self,
            level: Union[str, LogLevel] = LogLevel.INFO,
            format: str = "%(message)s",
            use_rich: bool = True,
            file_path: Optional[Path] = None,
            console_output: bool = True,
    ) -> None:
        self.level = LogLevel(level) if isinstance(level, str) else level
        self.format = format
        self.use_rich = use_rich
        self.file_path = file_path
        self.console_output = console_output


_loggers: Dict[str, Logger] = {}


def get_logger(
        name: Optional[str] = None,
        config: Optional[LogConfig] = None,
) -> Logger:
    """로거 인스턴스를 가져옵니다.

    Args:
        name: 로거 이름. None이면 root 로거 사용
        config: 로깅 설정. None이면 기본 설정 사용

    Returns:
        Logger 인터페이스를 구현한 로거 인스턴스
    """
    if name is None:
        name = "root"

    if name in _loggers:
        return _loggers[name]

    if config is None:
        config = LogConfig()

    logger = logging.getLogger(name)
    logger.setLevel(config.level.value)
    logger.handlers.clear()

    if config.console_output:
        if config.use_rich:
            console = Console(stderr=True)
            handler = RichHandler(
                console=console,
                rich_tracebacks=True,
                tracebacks_show_locals=True,
            )
            handler.setFormatter(logging.Formatter(config.format))
        else:
            handler = logging.StreamHandler(sys.stderr)
            handler.setFormatter(logging.Formatter(config.format))

        logger.addHandler(handler)

    if config.file_path:
        file_handler = logging.FileHandler(config.file_path, encoding="utf-8")
        file_handler.setFormatter(
            logging.Formatter(
                "%(asctime)s - %(name)s - %(levelname)s - %(message)s"
            )
        )
        logger.addHandler(file_handler)

    logger_impl = LoggerImpl(logger)
    _loggers[name] = logger_impl
    return logger_impl


def configure_root_logger(config: LogConfig) -> None:
    """루트 로거를 설정합니다.

    Args:
        config: 로깅 설정
    """
    logging.basicConfig(
        level=config.level.value,
        format=config.format,
        force=True,
    )

    if "root" in _loggers:
        del _loggers["root"]

    get_logger(config=config)
