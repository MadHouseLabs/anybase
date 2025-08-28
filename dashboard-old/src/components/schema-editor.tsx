import React, { useState, useEffect } from 'react';
import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { Textarea } from '@/components/ui/textarea';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { AlertCircle, Copy, CheckCircle } from 'lucide-react';
import { useToast } from '@/hooks/use-toast';

interface Schema {
  type: string;
  properties?: Record<string, any>;
  required?: string[];
  additionalProperties?: boolean | Record<string, any>;
  description?: string;
}

interface SchemaEditorProps {
  schema?: Schema;
  onChange: (schema: Schema) => void;
  readOnly?: boolean;
}

const defaultSchema: Schema = {
  type: 'object',
  properties: {},
  required: [],
  additionalProperties: true,
};

const exampleSchema = {
  type: 'object',
  description: 'User profile schema',
  properties: {
    name: {
      type: 'string',
      description: 'Full name',
      minLength: 1,
      maxLength: 100,
    },
    email: {
      type: 'string',
      format: 'email',
      description: 'Email address',
    },
    age: {
      type: 'integer',
      minimum: 0,
      maximum: 150,
      description: 'Age in years',
    },
    active: {
      type: 'boolean',
      default: true,
    },
    tags: {
      type: 'array',
      items: {
        type: 'string',
      },
      uniqueItems: true,
    },
  },
  required: ['name', 'email'],
  additionalProperties: false,
};

export const SchemaEditor: React.FC<SchemaEditorProps> = ({
  schema = defaultSchema,
  onChange,
  readOnly = false,
}) => {
  const { toast } = useToast();
  const [jsonValue, setJsonValue] = useState<string>('');
  const [jsonError, setJsonError] = useState<string>('');
  const [isValid, setIsValid] = useState(true);

  useEffect(() => {
    // Format the schema as pretty JSON
    try {
      setJsonValue(JSON.stringify(schema || defaultSchema, null, 2));
      setJsonError('');
      setIsValid(true);
    } catch (e) {
      console.error('Failed to stringify schema:', e);
    }
  }, [schema]);

  const handleJsonChange = (value: string) => {
    setJsonValue(value);
    
    if (readOnly) return;
    
    // Try to parse and validate the JSON
    try {
      const parsed = JSON.parse(value);
      
      // Basic validation
      if (typeof parsed !== 'object' || !parsed.type) {
        setJsonError('Schema must be an object with a "type" field');
        setIsValid(false);
        return;
      }
      
      // If it's an object type, ensure properties field exists
      if (parsed.type === 'object' && !parsed.properties) {
        parsed.properties = {};
      }
      
      setJsonError('');
      setIsValid(true);
      onChange(parsed);
    } catch (e) {
      setJsonError((e as Error).message);
      setIsValid(false);
    }
  };

  const formatJson = () => {
    try {
      const parsed = JSON.parse(jsonValue);
      setJsonValue(JSON.stringify(parsed, null, 2));
      setJsonError('');
      setIsValid(true);
    } catch (e) {
      setJsonError('Cannot format invalid JSON');
      setIsValid(false);
    }
  };

  const loadExample = () => {
    if (readOnly) return;
    setJsonValue(JSON.stringify(exampleSchema, null, 2));
    handleJsonChange(JSON.stringify(exampleSchema, null, 2));
  };

  const copyToClipboard = async () => {
    try {
      await navigator.clipboard.writeText(jsonValue);
      toast({
        title: 'Copied!',
        description: 'Schema copied to clipboard',
      });
    } catch (e) {
      toast({
        title: 'Error',
        description: 'Failed to copy to clipboard',
        variant: 'destructive',
      });
    }
  };

  return (
    <Card className="p-6">
      <div className="space-y-4">
        {/* Header with actions */}
        <div className="flex items-center justify-between">
          <div>
            <h3 className="text-lg font-semibold">OpenAPI Schema Definition</h3>
            <p className="text-sm text-muted-foreground">
              Define your collection schema using OpenAPI 3.0 JSON Schema format
            </p>
          </div>
          <div className="flex items-center gap-2">
            {!readOnly && (
              <>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={loadExample}
                >
                  Load Example
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={formatJson}
                  disabled={!isValid}
                >
                  Format JSON
                </Button>
              </>
            )}
            <Button
              variant="outline"
              size="sm"
              onClick={copyToClipboard}
            >
              <Copy className="h-4 w-4 mr-2" />
              Copy
            </Button>
          </div>
        </div>

        {/* Validation status */}
        {!readOnly && (
          <div className="flex items-center gap-2">
            {isValid ? (
              <div className="flex items-center gap-2 text-sm text-green-600">
                <CheckCircle className="h-4 w-4" />
                <span>Valid schema</span>
              </div>
            ) : (
              <div className="flex items-center gap-2 text-sm text-red-600">
                <AlertCircle className="h-4 w-4" />
                <span>Invalid schema</span>
              </div>
            )}
          </div>
        )}

        {/* Error message */}
        {jsonError && !readOnly && (
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>
              {jsonError}
            </AlertDescription>
          </Alert>
        )}

        {/* JSON editor */}
        <div className="relative">
          <Textarea
            value={jsonValue}
            onChange={(e) => handleJsonChange(e.target.value)}
            className="font-mono text-sm min-h-[400px] resize-y"
            placeholder="Enter OpenAPI schema JSON..."
            readOnly={readOnly}
            spellCheck={false}
          />
        </div>

        {/* Schema format help */}
        <div className="border rounded-lg p-4 bg-muted/30">
          <h4 className="text-sm font-semibold mb-2">Schema Format Reference</h4>
          <div className="grid grid-cols-2 gap-4 text-sm text-muted-foreground">
            <div>
              <p className="font-medium text-foreground mb-1">Common Types:</p>
              <ul className="space-y-1">
                <li>• string, number, integer, boolean</li>
                <li>• object (with properties)</li>
                <li>• array (with items)</li>
              </ul>
            </div>
            <div>
              <p className="font-medium text-foreground mb-1">String Formats:</p>
              <ul className="space-y-1">
                <li>• date-time, date, time</li>
                <li>• email, uri, uuid</li>
                <li>• ipv4, ipv6</li>
              </ul>
            </div>
            <div>
              <p className="font-medium text-foreground mb-1">Validation:</p>
              <ul className="space-y-1">
                <li>• minLength, maxLength (string)</li>
                <li>• minimum, maximum (number)</li>
                <li>• pattern (regex for string)</li>
              </ul>
            </div>
            <div>
              <p className="font-medium text-foreground mb-1">Other:</p>
              <ul className="space-y-1">
                <li>• required (array of field names)</li>
                <li>• enum (allowed values)</li>
                <li>• default (default value)</li>
              </ul>
            </div>
          </div>
        </div>
      </div>
    </Card>
  );
};