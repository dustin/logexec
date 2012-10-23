logexec is a simple command that runs a command and sends its stdout
to syslog at `INFO` level and its stderr to syslog at `ERROR` level.

Example:

    logexec -tag=hello echo hi

Causes the following to be logged at `INFO` level.

    Oct 22 23:18:13 myhostname hello:  hi
