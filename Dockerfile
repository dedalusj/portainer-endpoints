FROM alpine:3.5
ADD portainer-endpoints /
CMD ["/portainer-endpoints"]