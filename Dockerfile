FROM scratch
ADD ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
ADD crl_rotator /
CMD ["/crl_rotator"]
