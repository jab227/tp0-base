import signal


class SignalSIGTERM(Exception):
    def __init__(self, signum):
        self.signal = signal.Signals(signum).name


def sigterm_handler(signum, frame):
    _ = frame
    raise SignalSIGTERM(signum)


def sigterm_handler_init():
    signal.signal(signal.SIGTERM, sigterm_handler)
