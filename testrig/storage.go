/*
   GoToSocial
   Copyright (C) 2021 GoToSocial Authors admin@gotosocial.org

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package testrig

import (
	"fmt"
	"os"

	"github.com/superseriousbusiness/gotosocial/internal/storage"
)

// NewTestStorage returns a new in memory storage with the default test config
func NewTestStorage() storage.Storage {
	s, err := storage.NewInMem(NewTestConfig(), NewTestLog())
	if err != nil {
		panic(err)
	}
	return s
}

// StandardStorageSetup populates the storage with standard test entries from the given directory.
func StandardStorageSetup(s storage.Storage, relativePath string) {
	storedA := NewTestStoredAttachments()
	a := NewTestAttachments()
	for k, paths := range storedA {
		attachmentInfo, ok := a[k]
		if !ok {
			panic(fmt.Errorf("key %s not found in test attachments", k))
		}
		filenameOriginal := paths.original
		filenameSmall := paths.small
		pathOriginal := attachmentInfo.File.Path
		pathSmall := attachmentInfo.Thumbnail.Path
		bOriginal, err := os.ReadFile(fmt.Sprintf("%s/%s", relativePath, filenameOriginal))
		if err != nil {
			panic(err)
		}
		if err := s.StoreFileAt(pathOriginal, bOriginal); err != nil {
			panic(err)
		}
		bSmall, err := os.ReadFile(fmt.Sprintf("%s/%s", relativePath, filenameSmall))
		if err != nil {
			panic(err)
		}
		if err := s.StoreFileAt(pathSmall, bSmall); err != nil {
			panic(err)
		}
	}

	storedE := NewTestStoredEmoji()
	e := NewTestEmojis()
	for k, paths := range storedE {
		emojiInfo, ok := e[k]
		if !ok {
			panic(fmt.Errorf("key %s not found in test emojis", k))
		}
		filenameOriginal := paths.original
		filenameStatic := paths.static
		pathOriginal := emojiInfo.ImagePath
		pathStatic := emojiInfo.ImageStaticPath
		bOriginal, err := os.ReadFile(fmt.Sprintf("%s/%s", relativePath, filenameOriginal))
		if err != nil {
			panic(err)
		}
		if err := s.StoreFileAt(pathOriginal, bOriginal); err != nil {
			panic(err)
		}
		bStatic, err := os.ReadFile(fmt.Sprintf("%s/%s", relativePath, filenameStatic))
		if err != nil {
			panic(err)
		}
		if err := s.StoreFileAt(pathStatic, bStatic); err != nil {
			panic(err)
		}
	}
}

// StandardStorageTeardown deletes everything in storage so that it's clean for the next test
func StandardStorageTeardown(s storage.Storage) {
	keys, err := s.ListKeys()
	if err != nil {
		panic(err)
	}
	for _, k := range keys {
		if err := s.RemoveFileAt(k); err != nil {
			panic(err)
		}
	}
}
