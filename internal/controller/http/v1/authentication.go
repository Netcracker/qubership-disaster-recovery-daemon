package v1

import (
	"context"
	"fmt"
	"github.com/Netcracker/qubership-disaster-recovery-daemon/config"
	v1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"log"
	"net/http"
	"strings"
)

type Authenticator interface {
	CheckAuth(r *http.Request) (bool, string)
}

func NewTokenReviewAuthenticator(clientSet *kubernetes.Clientset, cfg config.AuthConfig) *TokenReviewAuthenticator {
	return &TokenReviewAuthenticator{
		clientSet: clientSet,
		config:    cfg,
	}
}

type TokenReviewAuthenticator struct {
	clientSet *kubernetes.Clientset
	config    config.AuthConfig
}

func (tra TokenReviewAuthenticator) CheckAuth(r *http.Request) (bool, string) {
	if !tra.config.AuthEnabled {
		return true, ""
	}
	token := tra.extractToken(r)
	if token == "" {
		return false, ""
	}
	reviewer := tra.clientSet.AuthenticationV1().TokenReviews()
	tokenReview := v1.TokenReview{}
	tokenReview.Spec.Token = token
	if tra.config.SiteManagerCustomAudience != "" {
		tokenReview.Spec.Audiences = []string{tra.config.SiteManagerCustomAudience}
	}
	reviewResult, err := reviewer.Create(context.TODO(), &tokenReview, metav1.CreateOptions{})
	if err != nil {
		log.Println("Can not create Kubernetes token reviewer")
		return false, token
	} else {
		authenticated := reviewResult.Status.Authenticated
		if authenticated {
			authenticated = tra.checkNamespaceAndSa(reviewResult)
		} else {
			log.Println("Unauthorized access")
		}
		return authenticated, token
	}
}

func (tra TokenReviewAuthenticator) extractToken(r *http.Request) string {
	bearerToken := r.Header.Get("Authorization")
	strArr := strings.Split(bearerToken, " ")
	if len(strArr) == 2 {
		return strArr[1]
	}
	return ""
}

func (tra TokenReviewAuthenticator) checkNamespaceAndSa(reviewResult *v1.TokenReview) bool {
	username := reviewResult.Status.User.Username
	authenticated := username == fmt.Sprintf("system:serviceaccount:%s:%s",
		tra.config.SiteManagerNamespace, tra.config.SiteManagerServiceAccountName)
	if !authenticated {
		log.Println("service account name or namespace of given token is not allowed")
	}
	return authenticated
}
