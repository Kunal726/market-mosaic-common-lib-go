package repositories

import (
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// BaseRepository interface defines common database operations
type BaseRepository[T any, ID any] interface {
	FindAll(logger *zap.Logger) ([]T, error)
	FindByID(logger *zap.Logger, id ID) (*T, error)
	Save(logger *zap.Logger, entity *T) error
	Delete(logger *zap.Logger, id ID) error
}

// BaseRepositoryImpl implements BaseRepository
type BaseRepositoryImpl[T any, ID any] struct {
	db *gorm.DB
}

func NewBaseRepository[T any, ID any](db *gorm.DB) BaseRepository[T, ID] {
	return &BaseRepositoryImpl[T, ID]{
		db: db,
	}
}

func (r *BaseRepositoryImpl[T, ID]) FindAll(logger *zap.Logger) ([]T, error) {
	logger.Info("Finding all entities")

	var entities []T
	err := r.db.Find(&entities).Error
	if err != nil {
		logger.Error("Failed to find all entities", zap.Error(err))
		return nil, err
	}

	logger.Info("Successfully found all entities",
		zap.Int("count", len(entities)))
	return entities, nil
}

func (r *BaseRepositoryImpl[T, ID]) FindByID(logger *zap.Logger, id ID) (*T, error) {
	logger.Info("Finding entity by ID", zap.Any("id", id))

	var entity T
	err := r.db.First(&entity, id).Error
	if err != nil {
		logger.Error("Failed to find entity by ID",
			zap.Error(err),
			zap.Any("id", id))
		return nil, err
	}

	logger.Info("Successfully found entity by ID", zap.Any("id", id))
	return &entity, nil
}

func (r *BaseRepositoryImpl[T, ID]) Save(logger *zap.Logger, entity *T) error {
	logger.Info("Saving entity")

	if err := r.db.Save(entity).Error; err != nil {
		logger.Error("Failed to save entity", zap.Error(err))
		return err
	}

	logger.Info("Successfully saved entity")
	return nil
}

func (r *BaseRepositoryImpl[T, ID]) Delete(logger *zap.Logger, id ID) error {
	logger.Info("Deleting entity", zap.Any("id", id))

	var entity T
	if err := r.db.Delete(&entity, id).Error; err != nil {
		logger.Error("Failed to delete entity",
			zap.Error(err),
			zap.Any("id", id))
		return err
	}

	logger.Info("Successfully deleted entity", zap.Any("id", id))
	return nil
}
