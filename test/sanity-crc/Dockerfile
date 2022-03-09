from registry.access.redhat.com/ubi8/ubi

RUN yum update -y && yum install -y sudo gettext

ADD . /
CMD ["./sanity-cli", "all"]
