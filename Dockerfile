FROM alpine

WORKDIR /root/mail2ics
COPY mail2ics /root/mail2ics
COPY config.json /root/mail2ics
CMD ./mail2ics
