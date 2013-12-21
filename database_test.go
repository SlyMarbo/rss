package rss

import (
	"testing"
)

var fixture = "Item.ID"

func TestDatabaseEnabled(t *testing.T) {
	if database.req <- fixture; <-database.res {
		t.Error("Should have not found the fixture string.")
	}
	if database.req <- fixture; !<-database.res {
		t.Error("Should have found the fixture string.")
	}
}

func TestDatabaseDisabled(t *testing.T) {
	CacheParsedItemIDs(false)

	if database.req <- fixture; <-database.res {
		t.Error("Should have not found the fixture string even though it was recorded.")
	}

	n := len(database.known)
	if database.req <- "foo"; <-database.res || len(database.known) != n {
		t.Error("Should not record a new entry.")
	}
}

func TestDatabaseReenabled(t *testing.T) {
	CacheParsedItemIDs(true)

	if database.req <- fixture; !<-database.res {
		t.Error("Should have found the fixture string again.")
	}

	n := len(database.known)
	if database.req <- "foo"; <-database.res || len(database.known) != n+1 {
		t.Error("Should record a new entry.")
	}

	if database.req <- "foo"; !<-database.res {
		t.Error("Should find the new entry.")
	}
}
