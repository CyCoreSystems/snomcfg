# snomcfg

`snomcfg` is a simple webserver which serves SNOM 3xx configuration files.

All files are expected to be created with one setting per line.  Each setting is
a key-value pair separated by a `:`.  Comments may be used be beginning any line
with `#`.

Default values:

 - `-l` Listen Address: `:8080`
 - `-d` Source directory: `/etc/asterisk/snom`

## `/config`

It is expected that the phone will supply its MAC address with the `?mac={mac}`
suffix.  Thus a complete URL may look like:

```
  http://phone.mydomain.com:8080/config?mac={mac}
```

For any given phone, configuration variables are read from the following list of
files, in order.  If any given setting exists in a later file, it will override any
previous value.

  - `snom-passwd.cfg`
  - `snomXXX.cfg` (where XXX is the phone's model number, such as 360)
  - `snom-001122334455.cfg (where 001122334455 is the phone's MAC address)`

## `/firmware`

Firmware reads firmware settings from the file `snomXXX-firmware.cfg` where XXX
is the phone's model number, such as 360.


