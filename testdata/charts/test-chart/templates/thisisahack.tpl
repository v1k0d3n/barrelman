{{ $ca := genCA "foo-ca" 365 }}
echo "{{$ca.Cert}}" > CA.cert;
echo "{{$ca.Key}}" > CA.key
