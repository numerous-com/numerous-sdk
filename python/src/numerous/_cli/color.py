from enum import Enum


class Color(Enum):
    BLACK = 30
    RED = 31
    GREEN = 32
    YELLOW = 33
    BLUE = 34
    MAGENTA = 35
    CYAN = 36
    WHITE = 37


def colorize(color: Color, message: str) -> str:
    return f"\033[;{color.value}m{message}\033[0m"


def yellow(message: str) -> str:
    return colorize(Color.YELLOW, message)


def green(message: str) -> str:
    return colorize(Color.GREEN, message)


def red(message: str) -> str:
    return colorize(Color.RED, message)


def black(message: str) -> str:
    return colorize(Color.BLACK, message)


def white(message: str) -> str:
    return colorize(Color.WHITE, message)


def blue(message: str) -> str:
    return colorize(Color.BLUE, message)


def magenta(message: str) -> str:
    return colorize(Color.MAGENTA, message)


def cyan(message: str) -> str:
    return colorize(Color.CYAN, message)
