package database

import (
	"gorm.io/gorm"
	"gorm.io/driver/sqlite"
)

// SlackThread represents a table with slackThread and threadSlug as composite primary key
type SlackThreadToSlug struct {
	SlackThread string `gorm:"primaryKey"`
	ThreadSlug  string
}

// Database interface abstracts database operations
type Interface interface {
	AutoMigrate() error
	CreateSlackThreadWithSlug(thread string, slug string) error
	GetSlugForThread(slackThread string) (string,bool, error)
	Close() error
}

// Database implements Database using gorm and sqlite
type Database struct {
	db *gorm.DB
}

// NewDatabase initializes a new sqlite database connection
func NewDatabase(dsn string) (*Database, error) {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return &Database{db: db}, nil
}

// AutoMigrate migrates the SlackThread schema
func (g *Database) AutoMigrate() error {
	return g.db.AutoMigrate(&SlackThreadToSlug{})
}

// CreateSlackThreadWithSlug inserts a new SlackThread record
func (g *Database) CreateSlackThreadWithSlug(thread string, slug string) error {
	return g.db.Create(&SlackThreadToSlug{SlackThread: thread, ThreadSlug: slug}).Error
}

// GetSlackThread retrieves a SlackThread by composite key
func (g *Database) GetSlugForThread(slackThread string) (string,bool, error) {
	var thread SlackThreadToSlug
	result := g.db.First(&thread, "slack_thread = ?", slackThread)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return "", false, nil
		}
		return "",false, result.Error
	}
	return thread.ThreadSlug, true, nil
}

// Close closes the database connection (noop for gorm v2, but included for interface)
func (g *Database) Close() error {
	sqlDB, err := g.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
