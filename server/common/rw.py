def send_all(sock, data: bytes):
    sent = 0
    while sent < len(data):
        n = sock.send(data[sent:])
        if n == 0:
            raise RuntimeError("connection was closed")
        sent += n

        
def recv_all(sock, read_size: int) -> bytes:
    buf = []
    read = 0
    while read < read_size:
        data = sock.recv(read_size - read)
        if data == b'':
            raise RuntimeError("connection was closed")
        read += len(data)
        buf.append(data)
    return b''.join(buf)
