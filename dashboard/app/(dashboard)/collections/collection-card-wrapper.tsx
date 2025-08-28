"use client";

import { useRouter } from "next/navigation";
import { ArrowRight } from "lucide-react";
import { Button } from "@/components/ui/button";
import { DeleteCollectionButton } from "./collection-client-components";
import { CollectionCardDisplay } from "./collection-card-server";

interface Collection {
  name: string;
  description?: string;
  settings?: {
    versioning?: boolean;
    soft_delete?: boolean;
    auditing?: boolean;
  };
  document_count?: number;
}

interface CollectionCardWrapperProps {
  collection: Collection;
}

export function CollectionCardWrapper({ collection }: CollectionCardWrapperProps) {
  const router = useRouter();

  return (
    <div onClick={() => router.push(`/collections/${collection.name}`)}>
      <CollectionCardDisplay collection={collection}>
        <Button
          variant="ghost"
          size="icon"
          className="h-10 w-10"
          onClick={(e) => {
            e.stopPropagation();
            router.push(`/collections/${collection.name}`);
          }}
        >
          <ArrowRight className="h-4 w-4" />
        </Button>
        <DeleteCollectionButton collectionName={collection.name} />
      </CollectionCardDisplay>
    </div>
  );
}