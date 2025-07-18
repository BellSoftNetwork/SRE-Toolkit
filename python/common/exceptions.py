"""공용 예외 클래스"""


class SREToolkitError(Exception):
    """SRE Toolkit 기본 예외"""

    pass


class ConfigError(SREToolkitError):
    """설정 관련 예외"""

    pass


class ValidationError(SREToolkitError):
    """검증 실패 예외"""

    pass


class NetworkError(SREToolkitError):
    """네트워크 관련 예외"""

    pass


class ResourceNotFoundError(SREToolkitError):
    """리소스를 찾을 수 없음"""

    pass


class TimeoutError(SREToolkitError):
    """타임아웃 예외"""

    pass
