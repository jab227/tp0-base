FROM python:3.9.7-slim
WORKDIR echoserver
COPY server .
RUN python -m unittest tests/test_common.py
ENTRYPOINT ["/bin/sh"]
