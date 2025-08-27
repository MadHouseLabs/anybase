// Migration script to remove roles collection and update users
// Run with: mongo localhost:27017/anybase scripts/migrate_remove_roles.js

// Update all users to have a single role instead of roles array
db.users.find({}).forEach(function(user) {
    var newRole = "developer"; // Default role
    
    // Check if user had admin role
    if (user.roles && user.roles.includes("admin")) {
        newRole = "admin";
    }
    
    // Update user with single role field
    db.users.updateOne(
        { _id: user._id },
        {
            $set: { 
                role: newRole,
                user_type: user.user_type || "regular"
            },
            $unset: { roles: "" }
        }
    );
    
    print("Updated user " + user.email + " with role: " + newRole);
});

// Drop the roles collection as we no longer need it
db.roles.drop();
print("Dropped roles collection");

// Drop permissions collection if it exists
if (db.getCollectionNames().indexOf("permissions") > -1) {
    db.permissions.drop();
    print("Dropped permissions collection");
}

print("Migration completed successfully!");