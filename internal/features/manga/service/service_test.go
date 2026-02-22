package service

import (
	"slices"
	"testing"

	"github.com/mairuu/mp-api/internal/features/manga/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessCoverArtChanges(t *testing.T) {
	s := &Service{}

	t.Run("adding single", func(t *testing.T) {
		existing := []model.CoverArt{}
		dtos := &[]UpdateCoverArtDTO{
			{Volume: "1", ObjectName: "staging-obj-1", Description: "Cover for volume 1"},
		}

		result, err := s.processCoverArtChanges(existing, dtos)
		require.NoError(t, err)

		assert.Len(t, result.Added, 1)
		assert.Equal(t, "1", result.Added[0].Volume)
		assert.Equal(t, "staging-obj-1", result.Added[0].ObjectName)

		assert.Len(t, result.Updated, 0)
		assert.Len(t, result.Deleted, 0)
	})

	t.Run("empty volumes", func(t *testing.T) {
		existing := []model.CoverArt{
			{Volume: "", ObjectName: "perma-obj-1", Description: "existing cover with empty volume"},
			{Volume: "", ObjectName: "perma-obj-2", Description: "deleting cover with empty volume"},
		}
		dtos := &[]UpdateCoverArtDTO{
			{Volume: "", ObjectName: "perma-obj-1", Description: "existing cover with empty volume"},
			{Volume: "", ObjectName: "staging-obj-3", Description: "cover with empty volume"},
			{Volume: "", ObjectName: "staging-obj-4", Description: "another cover with empty volume"},
		}

		result, err := s.processCoverArtChanges(existing, dtos)
		require.NoError(t, err)

		assert.Len(t, result.Added, 2)
		assert.True(t, slices.ContainsFunc(result.Added, func(c *model.CoverArt) bool {
			return c.ObjectName == "staging-obj-3"
		}))
		assert.True(t, slices.ContainsFunc(result.Added, func(c *model.CoverArt) bool {
			return c.ObjectName == "staging-obj-4"
		}))

		assert.Len(t, result.Updated, 1)
		assert.Equal(t, "perma-obj-1", result.Updated[0].ObjectName)

		assert.Len(t, result.Deleted, 1)
		assert.Equal(t, "perma-obj-2", result.Deleted[0].ObjectName)
	})

	t.Run("replace single (obj changed)", func(t *testing.T) {
		existing := []model.CoverArt{
			{Volume: "1", ObjectName: "perma-obj-1", Description: "Old cover"},
		}
		dtos := &[]UpdateCoverArtDTO{
			{Volume: "1", ObjectName: "staging-obj-2", Description: "New cover"},
		}

		result, err := s.processCoverArtChanges(existing, dtos)
		require.NoError(t, err)

		assert.Len(t, result.Added, 1)
		assert.Equal(t, "1", result.Added[0].Volume)
		assert.Equal(t, "staging-obj-2", result.Added[0].ObjectName)

		assert.Len(t, result.Updated, 0)

		assert.Len(t, result.Deleted, 1)
		assert.Equal(t, "perma-obj-1", result.Deleted[0].ObjectName)
	})

	t.Run("update single (obj unchanged)", func(t *testing.T) {
		existing := []model.CoverArt{
			{Volume: "1", ObjectName: "perma-obj-1", Description: "Old desc"},
		}
		dtos := &[]UpdateCoverArtDTO{
			{Volume: "1", ObjectName: "perma-obj-1", Description: "Updated desc"},
		}

		result, err := s.processCoverArtChanges(existing, dtos)
		require.NoError(t, err)

		assert.Len(t, result.Added, 0)

		assert.Len(t, result.Updated, 1)
		assert.Equal(t, "1", result.Updated[0].Volume)
		assert.Equal(t, "perma-obj-1", result.Updated[0].ObjectName)
		assert.Equal(t, "Updated desc", result.Updated[0].Description)

		assert.Len(t, result.Deleted, 0)
	})

	t.Run("delete single", func(t *testing.T) {
		existing := []model.CoverArt{
			{Volume: "1", ObjectName: "perma-obj-1", Description: "Cover"},
		}
		dtos := &[]UpdateCoverArtDTO{}

		result, err := s.processCoverArtChanges(existing, dtos)
		require.NoError(t, err)

		assert.Len(t, result.Added, 0)
		assert.Len(t, result.Updated, 0)

		assert.Len(t, result.Deleted, 1)
		assert.Equal(t, "1", result.Deleted[0].Volume)
		assert.Equal(t, "perma-obj-1", result.Deleted[0].ObjectName)
	})

	t.Run("mixed operations", func(t *testing.T) {
		existing := []model.CoverArt{
			{Volume: "", ObjectName: "perma-obj-0", Description: "Cover with empty volume"},
			{Volume: "1", ObjectName: "perma-obj-1", Description: "Cover 1"},
			{Volume: "2", ObjectName: "perma-obj-2", Description: "Cover 2"},
			{Volume: "3", ObjectName: "perma-obj-6", Description: "Cover 3"},
		}
		dtos := &[]UpdateCoverArtDTO{
			{Volume: "1", ObjectName: "perma-obj-1", Description: "Cover 1"},       // unchanged
			{Volume: "2", ObjectName: "staging-obj-2", Description: "New Cover 2"}, // replaced
			{Volume: "4", ObjectName: "staging-obj-4", Description: "New Cover 4"}, // added
			{Volume: "5", ObjectName: "perma-obj-6", Description: "Cover 5"},       // updated (same object, volume changed)
		}

		result, err := s.processCoverArtChanges(existing, dtos)
		require.NoError(t, err)

		// added: new volume 4 + replaced volume 2
		assert.Len(t, result.Added, 2)
		var addedVol4, addedVol2 *model.CoverArt
		for _, c := range result.Added {
			switch c.Volume {
			case "2":
				addedVol2 = c
			case "4":
				addedVol4 = c
			}
		}
		require.NotNil(t, addedVol2)
		assert.Equal(t, "staging-obj-2", addedVol2.ObjectName)
		require.NotNil(t, addedVol4)
		assert.Equal(t, "staging-obj-4", addedVol4.ObjectName)

		// updated: volume 1 (unchanged) and 3 (same object, volume changed)
		assert.Len(t, result.Updated, 2)
		var updatedVol1, updatedVol3 *model.CoverArt
		for _, c := range result.Updated {
			switch c.Volume {
			case "1":
				updatedVol1 = c
			case "5":
				updatedVol3 = c
			}
		}
		require.NotNil(t, updatedVol1)
		assert.Equal(t, "perma-obj-1", updatedVol1.ObjectName)
		require.NotNil(t, updatedVol3)
		assert.Equal(t, "perma-obj-6", updatedVol3.ObjectName)

		// deleted: old object for volume 2
		assert.Len(t, result.Deleted, 2)
		assert.True(t, slices.ContainsFunc(result.Deleted, func(c *model.CoverArt) bool {
			return c.ObjectName == "perma-obj-0"
		}))
		assert.True(t, slices.ContainsFunc(result.Deleted, func(c *model.CoverArt) bool {
			return c.ObjectName == "perma-obj-2"
		}))
	})

	t.Run("nil dtos returns existing as updated", func(t *testing.T) {
		existing := []model.CoverArt{
			{Volume: "1", ObjectName: "perma-obj-1"},
		}

		result, err := s.processCoverArtChanges(existing, nil)
		require.NoError(t, err)

		assert.Len(t, result.Added, 0)
		assert.Len(t, result.Updated, 1)
		assert.Len(t, result.Deleted, 0)
	})
}
