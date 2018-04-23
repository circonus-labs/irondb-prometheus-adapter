FROM scratch
ADD promadapter /promadapter
CMD ["/promadapter", "-log", "debug", "-addr", ":8080"]
