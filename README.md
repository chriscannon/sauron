# Sauron
A Go tool to concurrently read a file of newline separated IP addresses and count the number of IPs associated with a state using the GeoIP2 database.

## Example

Reading from stdin
```bash
cat ips_file | ./sauron -state="PA" -geoip="/usr/share/GeoIP/GeoIP2-City.mmdb"
```

Reading from a file
```bash
./sauron -state="PA" -geoip="/usr/share/GeoIP/GeoIP2-City.mmdb" -input="ips_file"
```

## Input
As input sauron expects a list of IPv4 IP addresses each on its own line. E.g.,
```
10.0.0.1
10.0.0.2
10.0.0.3
```
