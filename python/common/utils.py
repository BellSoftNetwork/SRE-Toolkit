"""공용 유틸리티 함수"""

import functools
import subprocess
import time
from dataclasses import dataclass
from datetime import datetime, timezone
from pathlib import Path
from typing import Any, Callable, Dict, List, Optional, TypeVar, Union

from .interfaces import Logger

T = TypeVar("T")


@dataclass
class CommandResult:
    """명령 실행 결과"""
    success: bool
    stdout: str
    stderr: str
    return_code: int


def retry(
        max_attempts: int = 3,
        delay: float = 1.0,
        backoff: float = 2.0,
        exceptions: tuple = (Exception,),
        logger: Optional[Logger] = None,
) -> Callable[[Callable[..., T]], Callable[..., T]]:
    """재시도 데코레이터

    Args:
        max_attempts: 최대 시도 횟수
        delay: 재시도 간 대기 시간 (초)
        backoff: 백오프 배수
        exceptions: 재시도할 예외 타입들
        logger: 로거 인스턴스

    Returns:
        데코레이터 함수
    """

    def decorator(func: Callable[..., T]) -> Callable[..., T]:
        @functools.wraps(func)
        def wrapper(*args: Any, **kwargs: Any) -> T:
            last_exception = None
            current_delay = delay

            for attempt in range(max_attempts):
                try:
                    return func(*args, **kwargs)
                except exceptions as e:
                    last_exception = e
                    if logger:
                        logger.warning(
                            f"{func.__name__} 실패 (시도 {attempt + 1}/{max_attempts}): {e}"
                        )

                    if attempt < max_attempts - 1:
                        time.sleep(current_delay)
                        current_delay *= backoff

            raise last_exception  # type: ignore

        return wrapper

    return decorator


def format_bytes(bytes_value: int, precision: int = 2) -> str:
    """바이트를 읽기 쉬운 형식으로 변환

    Args:
        bytes_value: 바이트 값
        precision: 소수점 자릿수

    Returns:
        포맷된 문자열
    """
    for unit in ["B", "KB", "MB", "GB", "TB", "PB"]:
        if bytes_value < 1024.0:
            return f"{bytes_value:.{precision}f} {unit}"
        bytes_value /= 1024.0
    return f"{bytes_value:.{precision}f} EB"


def format_duration(seconds: float) -> str:
    """초를 읽기 쉬운 시간 형식으로 변환

    Args:
        seconds: 초 단위 시간

    Returns:
        포맷된 문자열
    """
    if seconds < 60:
        return f"{seconds:.1f}초"
    elif seconds < 3600:
        minutes = int(seconds // 60)
        secs = int(seconds % 60)
        return f"{minutes}분 {secs}초"
    elif seconds < 86400:
        hours = int(seconds // 3600)
        minutes = int((seconds % 3600) // 60)
        return f"{hours}시간 {minutes}분"
    else:
        days = int(seconds // 86400)
        hours = int((seconds % 86400) // 3600)
        return f"{days}일 {hours}시간"


def get_timestamp(fmt: Optional[str] = None, utc: bool = True) -> str:
    """현재 타임스탬프를 가져옵니다

    Args:
        fmt: strftime 형식 문자열
        utc: UTC 시간 사용 여부

    Returns:
        포맷된 타임스탬프
    """
    now = datetime.now(timezone.utc) if utc else datetime.now()
    if fmt:
        return now.strftime(fmt)
    return now.isoformat()


def ensure_directory(path: Union[str, Path]) -> Path:
    """디렉토리가 존재하는지 확인하고 없으면 생성

    Args:
        path: 디렉토리 경로

    Returns:
        Path 객체
    """
    path = Path(path)
    path.mkdir(parents=True, exist_ok=True)
    return path


def merge_dicts(base: Dict[str, Any], *others: Dict[str, Any]) -> Dict[str, Any]:
    """여러 딕셔너리를 재귀적으로 병합

    Args:
        base: 기본 딕셔너리
        *others: 병합할 딕셔너리들

    Returns:
        병합된 딕셔너리
    """
    result = base.copy()

    for other in others:
        for key, value in other.items():
            if key in result and isinstance(result[key], dict) and isinstance(value, dict):
                result[key] = merge_dicts(result[key], value)
            else:
                result[key] = value

    return result


def flatten_dict(
        data: Dict[str, Any],
        parent_key: str = "",
        separator: str = ".",
) -> Dict[str, Any]:
    """중첩된 딕셔너리를 평탄화

    Args:
        data: 중첩된 딕셔너리
        parent_key: 부모 키
        separator: 키 구분자

    Returns:
        평탄화된 딕셔너리
    """
    items: List[tuple] = []

    for key, value in data.items():
        new_key = f"{parent_key}{separator}{key}" if parent_key else key

        if isinstance(value, dict):
            items.extend(flatten_dict(value, new_key, separator).items())
        else:
            items.append((new_key, value))

    return dict(items)


def run_command(
        command: List[str],
        timeout: Optional[int] = None,
        cwd: Optional[Union[str, Path]] = None,
        env: Optional[Dict[str, str]] = None,
) -> CommandResult:
    """외부 명령 실행

    Args:
        command: 실행할 명령어 리스트
        timeout: 타임아웃 (초)
        cwd: 작업 디렉토리
        env: 환경 변수

    Returns:
        CommandResult 객체
    """
    try:
        result = subprocess.run(
            command,
            capture_output=True,
            text=True,
            timeout=timeout,
            cwd=cwd,
            env=env,
        )

        return CommandResult(
            success=result.returncode == 0,
            stdout=result.stdout,
            stderr=result.stderr,
            return_code=result.returncode,
        )
    except subprocess.TimeoutExpired:
        return CommandResult(
            success=False,
            stdout="",
            stderr=f"Command timed out after {timeout} seconds",
            return_code=-1,
        )
    except Exception as e:
        return CommandResult(
            success=False,
            stdout="",
            stderr=str(e),
            return_code=-1,
        )
