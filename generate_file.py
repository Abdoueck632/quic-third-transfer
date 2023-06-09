import sys

PATTERN="ABCDEFGHIGKLMNOPQRSTUVWXYZ"
def main(size):
    with open(f"data_{size}.txt", "wt") as f:
        buf = ""
        for i in range(size):
            buf += PATTERN[i % len(PATTERN)]
        f.write(buf)

if __name__ == "__main__":
    main(int(sys.argv[1]))