package interceptors

import (
	"ClassConnectRPC/pkg/utils"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type ContextKey string

func AuthenticationInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

	// Skip some rpcs
	skipMethods := map[string]bool{
		"/main.ExecsService/Login":    true,
		"/main.ExecsService/AddExecs": true,
	}

	if skipMethods[info.FullMethod] {
		return handler(ctx, req)
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "Metadata unavailable")
	}

	authHeader, ok := md["authorization"]
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "Authorization token unavailable")
	}

	tokenStr := strings.TrimPrefix(authHeader[0], "Bearer ")
	tokenStr = strings.TrimSpace(tokenStr)

	// Check if token is blacklisted (user has logged out)
	isBlacklisted := utils.JwtStore.IsBlacklisted(tokenStr)
	if isBlacklisted {
		return nil, status.Error(codes.Unauthenticated, "Token has been revoked (logged out)")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		fmt.Println("ERROR: JWT_SECRET is not set")
		return nil, status.Error(codes.Internal, "JWT_SECRET not configured")
	}

	parsedToken, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
		_, ok := token.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			fmt.Printf("ERROR: Invalid signing method: %v\n", token.Method)
			return nil, status.Error(codes.Unauthenticated, "Invalid signing method")
		}
		return []byte(jwtSecret), nil
	})
	if err != nil {
		fmt.Printf("ERROR: Token parsing failed: %v\n", err)
		return nil, status.Error(codes.Unauthenticated, fmt.Sprintf("Token parsing failed: %v", err))
	}

	if !parsedToken.Valid {
		fmt.Println("ERROR: Token is not valid")
		return nil, status.Error(codes.Unauthenticated, "Token is not valid")
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok {
		fmt.Println("ERROR: Cannot parse claims")
		return nil, status.Error(codes.Unauthenticated, "Cannot parse claims")
	}

	role, ok := claims["role"].(string)
	if !ok {
		fmt.Printf("ERROR: Role claim missing or invalid. Claims: %v\n", claims)
		return nil, status.Error(codes.Unauthenticated, "Role claim missing")
	}

	userId, ok := claims["uid"].(string)
	if !ok {
		fmt.Printf("ERROR: UID claim missing or invalid. Claims: %v\n", claims)
		return nil, status.Error(codes.Unauthenticated, "UID claim missing")
	}

	username, ok := claims["user"].(string)
	if !ok {
		fmt.Printf("ERROR: User claim missing or invalid. Claims: %v\n", claims)
		return nil, status.Error(codes.Unauthenticated, "User claim missing")
	}

	expiresAtF64, ok := claims["exp"].(float64)
	if !ok {
		fmt.Printf("ERROR: Expiry claim missing or invalid. Claims: %v\n", claims)
		return nil, status.Error(codes.Unauthenticated, "Expiry claim missing")
	}
	expiresAtInt := int64(expiresAtF64)

	fmt.Printf("Authentication successful for user: %s (role: %s)\n", username, role)

	newCtx := context.WithValue(ctx, ContextKey("role"), role)
	newCtx = context.WithValue(newCtx, ContextKey("userId"), userId)
	newCtx = context.WithValue(newCtx, ContextKey("username"), username)
	newCtx = context.WithValue(newCtx, ContextKey("expiresAt"), expiresAtInt)

	return handler(newCtx, req)
}
