package core

import (
    "github.com/glerchundi/kubelistener/pkg/client/api/unversioned"
    "github.com/glerchundi/kubelistener/pkg/client/api/v1"
)

// Ingress is a collection of rules that allow inbound connections to reach the
// endpoints defined by a backend. An Ingress can be configured to give services
// externally-reachable urls, load balance traffic, terminate SSL, offer name
// based virtual hosting etc.
type Ingress struct {
    unversioned.TypeMeta `json:",inline"`
    // Standard object's metadata.
    // More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#metadata
    v1.ObjectMeta `json:"metadata,omitempty"`

    // Spec is the desired state of the Ingress.
    // More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#spec-and-status
    Spec IngressSpec `json:"spec,omitempty"`

    // Status is the current state of the Ingress.
    // More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#spec-and-status
    Status IngressStatus `json:"status,omitempty"`
}

// IngressList is a collection of Ingress.
type IngressList struct {
    unversioned.TypeMeta `json:",inline"`
    // Standard object's metadata.
    // More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#metadata
    unversioned.ListMeta `json:"metadata,omitempty"`

    // Items is the list of Ingress.
    Items []Ingress `json:"items"`
}

// IngressSpec describes the Ingress the user wishes to exist.
type IngressSpec struct {
    // A default backend capable of servicing requests that don't match any
    // rule. At least one of 'backend' or 'rules' must be specified. This field
    // is optional to allow the loadbalancer controller or defaulting logic to
    // specify a global default.
    Backend *IngressBackend `json:"backend,omitempty"`
    // A list of host rules used to configure the Ingress. If unspecified, or
    // no rule matches, all traffic is sent to the default backend.
    Rules []IngressRule `json:"rules,omitempty"`
    // TODO: Add the ability to specify load-balancer IP through claims
}

// IngressStatus describe the current state of the Ingress.
type IngressStatus struct {
    // LoadBalancer contains the current status of the load-balancer.
    LoadBalancer v1.LoadBalancerStatus `json:"loadBalancer,omitempty"`
}

// IngressRule represents the rules mapping the paths under a specified host to
// the related backend services. Incoming requests are first evaluated for a host
// match, then routed to the backend associated with the matching IngressRuleValue.
type IngressRule struct {
    // Host is the fully qualified domain name of a network host, as defined
    // by RFC 3986. Note the following deviations from the "host" part of the
    // URI as defined in the RFC:
    // 1. IPs are not allowed. Currently an IngressRuleValue can only apply to the
    //	  IP in the Spec of the parent Ingress.
    // 2. The `:` delimiter is not respected because ports are not allowed.
    //	  Currently the port of an Ingress is implicitly :80 for http and
    //	  :443 for https.
    // Both these may change in the future.
    // Incoming requests are matched against the host before the IngressRuleValue.
    // If the host is unspecified, the Ingress routes all traffic based on the
    // specified IngressRuleValue.
    Host string `json:"host,omitempty"`
    // IngressRuleValue represents a rule to route requests for this IngressRule.
    // If unspecified, the rule defaults to a http catch-all. Whether that sends
    // just traffic matching the host to the default backend or all traffic to the
    // default backend, is left to the controller fulfilling the Ingress. Http is
    // currently the only supported IngressRuleValue.
    IngressRuleValue `json:",inline,omitempty"`
}

// IngressRuleValue represents a rule to apply against incoming requests. If the
// rule is satisfied, the request is routed to the specified backend. Currently
// mixing different types of rules in a single Ingress is disallowed, so exactly
// one of the following must be set.
type IngressRuleValue struct {
    //TODO:
    // 1. Consider renaming this resource and the associated rules so they
    // aren't tied to Ingress. They can be used to route intra-cluster traffic.
    // 2. Consider adding fields for ingress-type specific global options
    // usable by a loadbalancer, like http keep-alive.

    HTTP *HTTPIngressRuleValue `json:"http,omitempty"`
}

// HTTPIngressRuleValue is a list of http selectors pointing to backends.
// In the example: http://<host>/<path>?<searchpart> -> backend where
// where parts of the url correspond to RFC 3986, this resource will be used
// to match against everything after the last '/' and before the first '?'
// or '#'.
type HTTPIngressRuleValue struct {
    // A collection of paths that map requests to backends.
    Paths []HTTPIngressPath `json:"paths"`
    // TODO: Consider adding fields for ingress-type specific global
    // options usable by a loadbalancer, like http keep-alive.
}

// HTTPIngressPath associates a path regex with a backend. Incoming urls matching
// the path are forwarded to the backend.
type HTTPIngressPath struct {
    // Path is a extended POSIX regex as defined by IEEE Std 1003.1,
    // (i.e this follows the egrep/unix syntax, not the perl syntax)
    // matched against the path of an incoming request. Currently it can
    // contain characters disallowed from the conventional "path"
    // part of a URL as defined by RFC 3986. Paths must begin with
    // a '/'. If unspecified, the path defaults to a catch all sending
    // traffic to the backend.
    Path string `json:"path,omitempty"`

    // Backend defines the referenced service endpoint to which the traffic
    // will be forwarded to.
    Backend IngressBackend `json:"backend"`
}

// IngressBackend describes all endpoints for a given service and port.
type IngressBackend struct {
    // Specifies the name of the referenced service.
    ServiceName string `json:"serviceName"`

    // Specifies the port of the referenced service.
    ServicePort v1.IntOrString `json:"servicePort"`
}