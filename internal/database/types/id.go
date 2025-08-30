package types

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	NilID = &genericID{} // Represents a nil/zero ID
)

// IDType represents the type of ID being used
type IDType string

const (
	IDTypeMongoDB  IDType = "mongodb"
	IDTypeUUID     IDType = "uuid"
)

// genericID implements the ID interface and can hold either ObjectID or UUID
type genericID struct {
	objectID *primitive.ObjectID
	uuid     *uuid.UUID
	idType   IDType
}

// NewID creates a new ID based on the database type
func NewID(dbType string) ID {
	switch dbType {
	case "mongodb":
		oid := primitive.NewObjectID()
		return &genericID{
			objectID: &oid,
			idType:   IDTypeMongoDB,
		}
	case "postgres":
		uid := uuid.New()
		return &genericID{
			uuid:   &uid,
			idType: IDTypeUUID,
		}
	default:
		// Default to UUID for unknown types
		uid := uuid.New()
		return &genericID{
			uuid:   &uid,
			idType: IDTypeUUID,
		}
	}
}

// ParseID parses a string into an ID
func ParseID(s string, dbType string) (ID, error) {
	if s == "" {
		return NilID, nil
	}

	switch dbType {
	case "mongodb":
		// Try to parse as ObjectID
		if len(s) == 24 {
			oid, err := primitive.ObjectIDFromHex(s)
			if err == nil {
				return &genericID{
					objectID: &oid,
					idType:   IDTypeMongoDB,
				}, nil
			}
		}
		return nil, fmt.Errorf("%w: invalid ObjectID format", ErrInvalidID)
		
	case "postgres":
		// Try to parse as UUID
		uid, err := uuid.Parse(s)
		if err == nil {
			return &genericID{
				uuid:   &uid,
				idType: IDTypeUUID,
			}, nil
		}
		return nil, fmt.Errorf("%w: invalid UUID format", ErrInvalidID)
		
	default:
		// Try UUID first, then ObjectID
		uid, err := uuid.Parse(s)
		if err == nil {
			return &genericID{
				uuid:   &uid,
				idType: IDTypeUUID,
			}, nil
		}
		
		if len(s) == 24 {
			oid, err := primitive.ObjectIDFromHex(s)
			if err == nil {
				return &genericID{
					objectID: &oid,
					idType:   IDTypeMongoDB,
				}, nil
			}
		}
		
		return nil, fmt.Errorf("%w: unknown format", ErrInvalidID)
	}
}

// FromObjectID creates an ID from a MongoDB ObjectID
func FromObjectID(oid primitive.ObjectID) ID {
	return &genericID{
		objectID: &oid,
		idType:   IDTypeMongoDB,
	}
}

// FromUUID creates an ID from a UUID
func FromUUID(uid uuid.UUID) ID {
	return &genericID{
		uuid:   &uid,
		idType: IDTypeUUID,
	}
}

// String returns the string representation of the ID
func (id *genericID) String() string {
	if id == nil || id.IsZero() {
		return ""
	}
	
	if id.objectID != nil {
		return id.objectID.Hex()
	}
	if id.uuid != nil {
		return id.uuid.String()
	}
	return ""
}

// Bytes returns the byte representation of the ID
func (id *genericID) Bytes() []byte {
	if id == nil || id.IsZero() {
		return nil
	}
	
	if id.objectID != nil {
		return id.objectID[:]
	}
	if id.uuid != nil {
		b, _ := id.uuid.MarshalBinary()
		return b
	}
	return nil
}

// IsZero returns true if the ID is zero/nil
func (id *genericID) IsZero() bool {
	if id == nil {
		return true
	}
	
	if id.objectID != nil {
		return id.objectID.IsZero()
	}
	if id.uuid != nil {
		return *id.uuid == uuid.Nil
	}
	return true
}

// Equals compares two IDs for equality
func (id *genericID) Equals(other ID) bool {
	if id == nil || other == nil {
		return id == other
	}
	
	otherGeneric, ok := other.(*genericID)
	if !ok {
		return false
	}
	
	if id.idType != otherGeneric.idType {
		return false
	}
	
	if id.objectID != nil && otherGeneric.objectID != nil {
		return *id.objectID == *otherGeneric.objectID
	}
	
	if id.uuid != nil && otherGeneric.uuid != nil {
		return *id.uuid == *otherGeneric.uuid
	}
	
	return false
}

// MarshalJSON implements json.Marshaler
func (id *genericID) MarshalJSON() ([]byte, error) {
	if id == nil || id.IsZero() {
		return json.Marshal(nil)
	}
	return json.Marshal(id.String())
}

// UnmarshalJSON implements json.Unmarshaler
func (id *genericID) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	
	if s == "" {
		return nil
	}
	
	// Try to determine the type from the format
	if len(s) == 24 {
		// Likely ObjectID
		oid, err := primitive.ObjectIDFromHex(s)
		if err == nil {
			id.objectID = &oid
			id.idType = IDTypeMongoDB
			return nil
		}
	}
	
	// Try UUID
	uid, err := uuid.Parse(s)
	if err == nil {
		id.uuid = &uid
		id.idType = IDTypeUUID
		return nil
	}
	
	return fmt.Errorf("%w: cannot unmarshal %s", ErrInvalidID, s)
}

// ToObjectID converts the ID to MongoDB ObjectID if possible
func (id *genericID) ToObjectID() (primitive.ObjectID, error) {
	if id.objectID != nil {
		return *id.objectID, nil
	}
	return primitive.NilObjectID, fmt.Errorf("ID is not an ObjectID")
}

// ToUUID converts the ID to UUID if possible
func (id *genericID) ToUUID() (uuid.UUID, error) {
	if id.uuid != nil {
		return *id.uuid, nil
	}
	return uuid.Nil, fmt.Errorf("ID is not a UUID")
}

// Type returns the type of the ID
func (id *genericID) Type() IDType {
	return id.idType
}

// ConvertObjectIDsToIDs converts a slice of ObjectIDs to generic IDs
func ConvertObjectIDsToIDs(oids []primitive.ObjectID) []ID {
	ids := make([]ID, len(oids))
	for i, oid := range oids {
		ids[i] = FromObjectID(oid)
	}
	return ids
}

// ConvertUUIDsToIDs converts a slice of UUIDs to generic IDs
func ConvertUUIDsToIDs(uuids []uuid.UUID) []ID {
	ids := make([]ID, len(uuids))
	for i, uid := range uuids {
		ids[i] = FromUUID(uid)
	}
	return ids
}

// ConvertIDsToObjectIDs converts generic IDs to ObjectIDs (returns error if any ID is not ObjectID)
func ConvertIDsToObjectIDs(ids []ID) ([]primitive.ObjectID, error) {
	oids := make([]primitive.ObjectID, len(ids))
	for i, id := range ids {
		gid, ok := id.(*genericID)
		if !ok || gid.objectID == nil {
			return nil, fmt.Errorf("ID at index %d is not an ObjectID", i)
		}
		oids[i] = *gid.objectID
	}
	return oids, nil
}

// ConvertIDsToUUIDs converts generic IDs to UUIDs (returns error if any ID is not UUID)
func ConvertIDsToUUIDs(ids []ID) ([]uuid.UUID, error) {
	uuids := make([]uuid.UUID, len(ids))
	for i, id := range ids {
		gid, ok := id.(*genericID)
		if !ok || gid.uuid == nil {
			return nil, fmt.Errorf("ID at index %d is not a UUID", i)
		}
		uuids[i] = *gid.uuid
	}
	return uuids, nil
}

// IsValidObjectIDString checks if a string is a valid ObjectID
func IsValidObjectIDString(s string) bool {
	if len(s) != 24 {
		return false
	}
	_, err := hex.DecodeString(s)
	return err == nil
}

// IsValidUUIDString checks if a string is a valid UUID
func IsValidUUIDString(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}

// GenerateIDForField generates an appropriate ID for a field that might contain "_id"
// This is useful for maintaining compatibility when switching between databases
func GenerateIDForField(field string, value interface{}, dbType string) (interface{}, error) {
	if field != "_id" && field != "id" {
		return value, nil
	}
	
	// If value is already an ID, return it
	if _, ok := value.(ID); ok {
		return value, nil
	}
	
	// If value is a string, try to parse it
	if s, ok := value.(string); ok {
		if s == "" {
			return NewID(dbType), nil
		}
		return ParseID(s, dbType)
	}
	
	// If value is ObjectID (MongoDB specific)
	if oid, ok := value.(primitive.ObjectID); ok {
		return FromObjectID(oid), nil
	}
	
	// If value is UUID
	if uid, ok := value.(uuid.UUID); ok {
		return FromUUID(uid), nil
	}
	
	// Generate new ID if value is nil
	if value == nil {
		return NewID(dbType), nil
	}
	
	return nil, fmt.Errorf("cannot convert value to ID: %v", value)
}