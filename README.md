gomaild2
--------
Hopefully the successor to [gomaild](https://github.com/trapped/gomaild).
Requires [`gengen`](https://github.com/trapped/gengen) to generate some files (`go get github.com/trapped/gengen`).

##Status

- [ ] DB interface to actually store emails
- [x] YAML config
- [x] Logging (logfmt)
- [ ] SMTP server
  - [x] `HELO`
  - [x] `EHLO` (implicit `PIPELINING` and `8BITMIME`)
  - [x] `NOOP`
  - [x] `MAIL FROM`
  - [x] `RCPT TO`
  - [x] `DATA`
  - [x] `QUIT`
  - [ ] `AUTH`
  - [ ] `STARTTLS`
- [ ] SMTP client (to send outbound emails)
- [ ] POP3 server
