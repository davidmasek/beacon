from dataclasses import dataclass
from datetime import datetime, timezone


@dataclass(frozen=True)
class Timestamp:
    dt: datetime

    @staticmethod
    def now():
        return Timestamp(datetime.now(timezone.utc))

    def format(self) -> str:
        # already timezone-aware we convert to UTC (possibly changing the time values)
        if self.dt.tzinfo:
            iso = self.dt.astimezone(timezone.utc).isoformat(
                sep="T", timespec="seconds"
            )
        # not timezone aware - assumme UTC without changing time values
        else:
            iso = self.dt.replace(tzinfo=timezone.utc).isoformat(
                sep="T", timespec="seconds"
            )
        # convert to expected format ending with Z
        return iso.replace("+00:00", "Z")
