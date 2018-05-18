FROM ubuntu:latest

# This is needed for flow, and the weirdos that built it in ocaml:
RUN apt-get update && apt-get install -y build-essential
