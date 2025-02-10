
#include <bpf/ctx/unspec.h>


static __always_inline __maybe_unused bool is_v4_loopback(__be32 daddr)
{
    // check 127.0.0.0/8 range
    return (daddr & bpf_htonl(0xff000000)) == bpf_htonl(0x7f000000)
}

static __always_inline __maybe_unused bool is_v6_loopback(__be32 daddr)
{
    // check ipv6 loopback
}

static __always_inline __maybe_unused __be16
ctx_dst_port(const struct bpf_sock_addr *ctx)
{
    // assign ctx userport to destination port
}

static __always_inline __maybe_unused __be16
ctx_src_port(const struct bpf_sock *ctx){
    // read the src port from context 
}

static __always_inline __maybe_unused
void ctx_set_port(struct bpf_sock_addr *ctx, __be16 dport) 
{
    // assign destination port to user port
}

static __always__inline __maybe_unused bool task_in_extended_hostns(void)
{
    // extension for non-dolphin managed containers
}

static __always_inline __maybe_unused bool 
ctx_in_hostns(void *ctx __maybe_unused, __net_cookie *cookie)
{
    // assign cookies
}

static __always_inline __maybe_unused
bool sock_is_health_check(struct bpf_sock_addr *ctx __maybe_unused)
{
        // check socket associaed to this ethernet interface
}

static __always_inline __maybe_unused
__u64 sock_select_slot(struct bpf_sock_addr *ctx)
{
    // setthe socket local cookie
}

static __alwas_inline __maybe_unused
bool sock_proto_enabled(__u32 proto)
{}
