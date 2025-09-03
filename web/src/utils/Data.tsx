import type {
  ResultData,
  ValMatrix,
  ValScalar,
  ValVector,
} from '@/types/Data.ts';

type ProcessedResult =
  | {
      type: 'scalar';
      processed: ValScalar;
      original: ResultData;
    }
  | {
      type: 'vector';
      processed: ValVector[];
      original: ResultData;
    }
  | {
      type: 'matrix';
      processed: ValMatrix[];
      original: ResultData;
    }
  | {
      type: 'unknown';
      processed: unknown;
      original: ResultData;
    };

export function processResult(result?: ResultData): ProcessedResult | null {
  if (!result) {
    console.log('Result Data cannot be empty');
    return null;
  }

  switch (result.type) {
    case 'scalar':
      const scalar = result.data as ValScalar;
      return { type: 'scalar', processed: scalar, original: result };

    case 'vector':
      const vector = result.data as ValVector[];
      return { type: 'vector', processed: vector, original: result };

    case 'matrix':
      const matrix = result.data as ValMatrix[];
      return { type: 'matrix', processed: matrix, original: result };

    default:
      console.log('Unknown result type:', result.type);
      return { type: 'unknown', processed: result.data, original: result };
  }
}
