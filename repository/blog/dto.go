package blog_repository

import (
	"time"

	"afikrim_a.bitbucket.org/simple-go-queue/core/entity"
)

const (
	TableName = "blog"
)

type BlogDto struct {
	ID        int64     `gorm:"column:id;primary_key"`
	Title     string    `gorm:"column:title"`
	Body      string    `gorm:"column:body"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (BlogDto) TableName() string {
	return TableName
}

func (dto *BlogDto) ToEntity() *entity.Blog {
	if dto == nil {
		return nil
	}

	return &entity.Blog{
		ID:    dto.ID,
		Title: dto.Title,
		Body:  dto.Body,
	}
}

func (BlogDto) FromEntity(entity *entity.Blog) *BlogDto {
	if entity == nil {
		return nil
	}

	return &BlogDto{
		ID:    entity.ID,
		Title: entity.Title,
		Body:  entity.Body,
	}
}

func (dto *BlogDto) InitTimestamps() *BlogDto {
	if dto == nil {
		return nil
	}

	now := time.Now()
	dto.CreatedAt = now
	dto.UpdatedAt = now
	return dto
}

func (dto *BlogDto) UpdateTimestamps() *BlogDto {
	if dto == nil {
		return nil
	}

	now := time.Now()
	dto.UpdatedAt = now
	return dto
}
