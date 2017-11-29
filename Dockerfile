FROM scratch
ADD ./cmd/cmd_linux /cmd
CMD [ "/cmd" ]