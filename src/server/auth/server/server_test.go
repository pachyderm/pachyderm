package server

import (
	"testing"

	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"golang.org/x/net/context"
)

func TestAuth(t *testing.T) {
	apiServer, err := NewAPIServer("localhost:32379")
	require.NoError(t, err)
	_, err = apiServer.PutActivationCode(context.Background(), &auth.ActivationCode{
		Token:     `{"expiry":"2017-07-08T18:11:11.362Z","scopes":{"basic":true},"name":"Pachyderm"}`,
		Signature: `HjvekyffCpzcTsGDCw/gtZjMY0x2bzAO4jLdDQNGPVeV81IBbqooDSJY6gzDNui82IxaKMGX6kvavP/erz458S5cevFUwJiCjTQywOSiog+aNbHcl4Km1uUEoJYEr4/mVH9qTqxUZnqYFKdzvJlwHKw7I17uVWRtQNlZqTOUPgjlqWiJMBVMalA92v898kgVYk/a/J5bidcMKsoLKAbmIC19Dn4Qlpj+pahwTgbarBLTos8BGT5/B4y6eFHBCJDBWbYa1jREb+Ibyonk8yPNXFbfFo3U6QDxfCBARLP9DvMwwtWyMhelqM4trn76+AWfx+ufedu054Qi/Ux2phHmwPo+gmJKHNHiRo5S/YvkoN9KchncV1ZHi3VVxSrKgVoFFONoh6Q4NaWIWDYnmQ8r8k+2PzWUKLryYeXpAm01J0z8U2P7jNtZg9L7x1v7pbwDGZk6CYof8ssIvQxkwGoIWM5JlVV1lZJ1MH1A1I30dG3G/LZxtcJuVRDAkgVS5gRQi0vU05aSOjGSj1cCvk5OyvTRsVYqn2Ks+Z/9HCqnTcdRjD7LP3oA5qm7vTzwLnC3XVQGDsrjLvKxSS/dwxaKqWExM9VMjppc8u8E/i3mRSMd8q4iKkOhRkotMgtqzARLwltTa53RdfNxh23yxK1Hszbcr3c9s6tSLI+xWWpnfUo=`,
	})
	require.NoError(t, err)
}
