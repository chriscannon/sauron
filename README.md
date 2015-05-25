# Sauron
A Go tool to concurrently read a file of newline separated IP addresses and count the number of IPs associated with a state using the GeoIP2 database.

## Example

Reading from stdin
```golang
cat ips_file | ./sauron -state="PA" -geoip="/usr/share/GeoIP/GeoIP2-City.mmdb"
```

Reading from a file
```golang
./sauron -state="PA" -geoip="/usr/share/GeoIP/GeoIP2-City.mmdb" -input="ips_file"
```
