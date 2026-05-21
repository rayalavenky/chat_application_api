package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID             primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	FirstName      string             `json:"firstName" bson:"firstName" binding:"required"`
	LastName       string             `json:"lastName" bson:"lastName" binding:"required"`
	PhoneNumber    string             `json:"phoneNumber" bson:"phoneNumber" binding:"required"`
	Email          string             `json:"email" bson:"email" binding:"required,email"`
	Age            int                `json:"age" bson:"age" binding:"required,gte=1"`
	Bio            string             `json:"bio" bson:"bio,omitempty"`
	Password       string             `json:"password,omitempty" bson:"password" binding:"required,min=6"`
	Role           string             `json:"role" bson:"role"`
	IsOnline       bool               `json:"isOnline" bson:"isOnline"`
	LastSeen       time.Time          `json:"lastSeen" bson:"lastSeen"`
	CreatedAt      time.Time          `json:"createdAt" bson:"createdAt"`
	ResetOTP       string             `bson:"resetOtp,omitempty"`
	ResetOTPExpiry time.Time          `bson:"resetOtpExpiry,omitempty"`
}

type RegisterRequest struct {
	FirstName   string `json:"firstName" binding:"required,min=3,max=10"`
	LastName    string `json:"lastName" binding:"required,min=3,max=10"`
	PhoneNumber string `json:"phoneNumber" binding:"required,min=10,max=10"`
	Email       string `json:"email" binding:"required,email"`
	Age         int    `json:"age" binding:"required,gte=18"`
}

type UserResponse struct {
	ID          primitive.ObjectID `json:"id"`
	FirstName   string             `json:"firstName"`
	LastName    string             `json:"lastName"`
	PhoneNumber string             `json:"phoneNumber"`
	Email       string             `json:"email"`
	Age         int                `json:"age"`
	Bio         string             `json:"bio"`
	IsOnline    bool               `json:"isOnline"`
	LastSeen    time.Time          `json:"lastSeen"`
	CreatedAt   time.Time          `json:"createdAt"`
}

type RefreshToken struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	UserID    primitive.ObjectID `json:"userId" bson:"userId"`
	Token     string             `json:"token" bson:"token"`
	ExpiresAt time.Time          `json:"expiresAt" bson:"expiresAt"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	ID           primitive.ObjectID `json:"id"`
	FirstName    string             `json:"firstName"`
	LastName     string             `json:"lastName"`
	PhoneNumber  string             `json:"phoneNumber"`
	Email        string             `json:"email"`
	Age          int                `json:"age"`
	Role         string             `json:"role"`
	IsOnline     bool               `json:"isOnline"`
	LastSeen     time.Time          `json:"lastSeen"`
	CreatedAt    time.Time          `json:"createdAt"`
	AccessToken  string             `json:"accessToken"`
	RefreshToken string             `json:"refreshToken"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type VerifyOTPRequest struct {
	Email string `json:"email" binding:"required,email"`
	OTP   string `json:"otp" binding:"required"`
}

type ResetPasswordRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type UpdateProfileRequest struct {
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	PhoneNumber string `json:"phoneNumber"`
	Age         int    `json:"age"`
	Bio         string `json:"bio"`
}

type UserSearchRequest struct {
	Search UserSearch `json:"search"`
}

type UserSearch struct {
	PhoneNumber string `json:"phoneNumber"`
	FirstName   string `json:"firstName"`
	LastName    string `json:"lastName"`
	Email       string `json:"email"`
}

type UserRequest struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	FromUserID primitive.ObjectID `bson:"fromUserId" json:"fromUserId"`
	ToUserID   primitive.ObjectID `bson:"toUserId" json:"toUserId"`
	Status     string             `bson:"status" json:"status"` // pending, accepted, rejected
	CreatedAt  time.Time          `bson:"createdAt" json:"createdAt"`
}

type Contact struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id"`

	UserID primitive.ObjectID `bson:"userId" json:"userId"`

	ContactUserID primitive.ObjectID `bson:"contactUserId" json:"contactUserId"`

	FirstName   string `bson:"firstName" json:"firstName"`
	LastName    string `bson:"lastName" json:"lastName"`
	Email       string `bson:"email" json:"email"`
	PhoneNumber string `bson:"phoneNumber" json:"phoneNumber"`

	IsOnline bool `bson:"isOnline" json:"isOnline"`

	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
}
