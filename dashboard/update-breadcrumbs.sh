#!/bin/bash

# List of pages to update with their names
declare -A pages=(
  ["users"]="Users"
  ["access-keys"]="Access Keys"
  ["views"]="Views"
  ["functions"]="Functions"
  ["storage"]="Storage"
  ["cron"]="Cron Jobs"
  ["realtime"]="Realtime"
  ["settings"]="Settings"
  ["api-docs"]="API Documentation"
)

for page in "${!pages[@]}"; do
  file="app/(dashboard)/${page}/page.tsx"
  name="${pages[$page]}"
  
  echo "Updating $file with breadcrumbs for $name..."
  
  # Add breadcrumb import if not present
  if ! grep -q "@/components/ui/breadcrumb" "$file"; then
    # Find the last import line and add breadcrumb import after it
    sed -i '' '/^import.*from/{ 
      $!{ 
        N
        s/\(.*\)\n/\1\nimport {\n  Breadcrumb,\n  BreadcrumbItem,\n  BreadcrumbLink,\n  BreadcrumbList,\n  BreadcrumbPage,\n  BreadcrumbSeparator,\n} from "@\/components\/ui\/breadcrumb";\n/
      }
    }' "$file"
  fi
  
  # Add breadcrumbs after container opening
  sed -i '' "s/<div className=\"container mx-auto py-8 max-w-7xl\">/<div className=\"container mx-auto py-8 max-w-7xl\">\n      {\/* Breadcrumbs *\/}\n      <Breadcrumb className=\"mb-6\">\n        <BreadcrumbList>\n          <BreadcrumbItem>\n            <BreadcrumbLink href=\"\/\">Dashboard<\/BreadcrumbLink>\n          <\/BreadcrumbItem>\n          <BreadcrumbSeparator \/>\n          <BreadcrumbItem>\n            <BreadcrumbPage>${name}<\/BreadcrumbPage>\n          <\/BreadcrumbItem>\n        <\/BreadcrumbList>\n      <\/Breadcrumb>\n/" "$file"
done

echo "All pages updated with breadcrumbs!"