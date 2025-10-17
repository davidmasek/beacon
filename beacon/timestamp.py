from dataclasses import dataclass
from datetime import datetime, timezone


@dataclass(frozen=True)
class Timestamp:
    dt: datetime

    @staticmethod
    def now():
        return Timestamp(datetime.now(timezone.utc))

    def format(self) -> str:
        iso = self.dt.astimezone(timezone.utc).isoformat(sep="T", timespec="seconds")
        return iso.replace("+00:00", "Z")
