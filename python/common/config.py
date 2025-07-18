"""설정 관리 공용 모듈"""

import json
import os
import yaml
from pathlib import Path
from typing import Any, Dict, List, Optional, Union

from .interfaces import ConfigLoader, ConfigProvider


class YAMLLoader:
    """YAML 파일 로더"""

    def load(self, path: Path) -> Dict[str, Any]:
        """YAML 파일을 로드합니다."""
        if not path.exists():
            raise FileNotFoundError(f"설정 파일을 찾을 수 없습니다: {path}")

        with open(path, "r", encoding="utf-8") as f:
            return yaml.safe_load(f) or {}

    def validate(self, data: Dict[str, Any]) -> bool:
        """YAML 데이터 검증 (기본 구현은 항상 True)"""
        return isinstance(data, dict)


class JSONLoader:
    """JSON 파일 로더"""

    def load(self, path: Path) -> Dict[str, Any]:
        """JSON 파일을 로드합니다."""
        if not path.exists():
            raise FileNotFoundError(f"설정 파일을 찾을 수 없습니다: {path}")

        with open(path, "r", encoding="utf-8") as f:
            return json.load(f)

    def validate(self, data: Dict[str, Any]) -> bool:
        """JSON 데이터 검증 (기본 구현은 항상 True)"""
        return isinstance(data, dict)


class Config(ConfigProvider):
    """설정 관리 클래스"""

    def __init__(self, data: Optional[Dict[str, Any]] = None) -> None:
        """Config 인스턴스를 초기화합니다.

        Args:
            data: 초기 설정 데이터
        """
        self._data = data or {}
        self._loaders: Dict[str, ConfigLoader] = {
            ".yaml": YAMLLoader(),
            ".yml": YAMLLoader(),
            ".json": JSONLoader(),
        }

    def load_file(self, path: Union[str, Path]) -> None:
        """파일에서 설정을 로드합니다.

        Args:
            path: 설정 파일 경로
        """
        path = Path(path)
        suffix = path.suffix.lower()

        if suffix not in self._loaders:
            raise ValueError(f"지원하지 않는 파일 형식입니다: {suffix}")

        loader = self._loaders[suffix]
        data = loader.load(path)

        if not loader.validate(data):
            raise ValueError("설정 데이터 검증에 실패했습니다")

        self._merge(data)

    def load_env(self, prefix: str = "") -> None:
        """환경 변수에서 설정을 로드합니다.

        Args:
            prefix: 환경 변수 접두사
        """
        for key, value in os.environ.items():
            if prefix and not key.startswith(prefix):
                continue

            if prefix:
                key = key[len(prefix):].lower()
            else:
                key = key.lower()

            # 중첩된 키 처리 (예: APP_DB_HOST -> app.db.host)
            keys = key.split("_")
            self._set_nested(keys, value)

    def get(self, key: str, default: Any = None) -> Any:
        """설정 값을 가져옵니다.

        Args:
            key: 설정 키 (점 표기법 지원)
            default: 기본값

        Returns:
            설정 값
        """
        keys = key.split(".")
        value = self._data

        for k in keys:
            if isinstance(value, dict) and k in value:
                value = value[k]
            else:
                return default

        return value

    def set(self, key: str, value: Any) -> None:
        """설정 값을 설정합니다.

        Args:
            key: 설정 키 (점 표기법 지원)
            value: 설정 값
        """
        keys = key.split(".")
        self._set_nested(keys, value)

    def exists(self, key: str) -> bool:
        """설정 키가 존재하는지 확인합니다.

        Args:
            key: 설정 키 (점 표기법 지원)

        Returns:
            키 존재 여부
        """
        keys = key.split(".")
        value = self._data

        for k in keys:
            if isinstance(value, dict) and k in value:
                value = value[k]
            else:
                return False

        return True

    def to_dict(self) -> Dict[str, Any]:
        """전체 설정을 딕셔너리로 반환합니다."""
        return self._data.copy()

    def _merge(self, data: Dict[str, Any]) -> None:
        """설정 데이터를 병합합니다."""
        self._data = self._deep_merge(self._data, data)

    def _deep_merge(self, base: Dict[str, Any], update: Dict[str, Any]) -> Dict[str, Any]:
        """딕셔너리를 재귀적으로 병합합니다."""
        result = base.copy()

        for key, value in update.items():
            if key in result and isinstance(result[key], dict) and isinstance(value, dict):
                result[key] = self._deep_merge(result[key], value)
            else:
                result[key] = value

        return result

    def _set_nested(self, keys: List[str], value: Any) -> None:
        """중첩된 키에 값을 설정합니다."""
        current = self._data

        for i, key in enumerate(keys[:-1]):
            if key not in current:
                current[key] = {}
            elif not isinstance(current[key], dict):
                raise ValueError(f"키 '{'.'.join(keys[:i + 1])}'는 딕셔너리가 아닙니다")
            current = current[key]

        # 환경 변수 값 타입 변환
        if isinstance(value, str):
            if value.lower() in ("true", "false"):
                value = value.lower() == "true"
            elif value.isdigit():
                value = int(value)
            else:
                try:
                    value = float(value)
                except ValueError:
                    pass

        current[keys[-1]] = value


def load_config(
        config_files: Optional[List[Union[str, Path]]] = None,
        env_prefix: Optional[str] = None,
        defaults: Optional[Dict[str, Any]] = None,
) -> Config:
    """설정을 로드하는 헬퍼 함수

    Args:
        config_files: 로드할 설정 파일 목록
        env_prefix: 환경 변수 접두사
        defaults: 기본 설정값

    Returns:
        Config 인스턴스
    """
    config = Config(defaults)

    if config_files:
        for file_path in config_files:
            try:
                config.load_file(file_path)
            except FileNotFoundError:
                pass

    if env_prefix:
        config.load_env(env_prefix)

    return config
