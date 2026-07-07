import { StrataAuthClient } from './auth';

export type FilterOperator = 'eq' | 'neq' | 'gt' | 'gte' | 'lt' | 'lte' | 'like' | 'is' | 'in';

export interface QueryFilter {
  column: string;
  operator: FilterOperator;
  value: any;
}

export class StrataRestClient<T = any> {
  private method: 'GET' | 'POST' | 'PATCH' | 'DELETE' = 'GET';
  private selectCols?: string;
  private filters: QueryFilter[] = [];
  private orderBy?: string;
  private limitVal?: number;
  private offsetVal?: number;
  private bodyData?: any;

  constructor(
    private url: string,
    private table: string,
    private auth: StrataAuthClient
  ) {}

  public select(columns: string = '*'): this {
    this.method = 'GET';
    this.selectCols = columns;
    return this;
  }

  public insert(data: any): this {
    this.method = 'POST';
    this.bodyData = data;
    return this;
  }

  public update(data: any): this {
    this.method = 'PATCH';
    this.bodyData = data;
    return this;
  }

  public delete(): this {
    this.method = 'DELETE';
    return this;
  }

  // Filters
  public eq(column: string, value: any): this {
    this.filters.push({ column, operator: 'eq', value });
    return this;
  }

  public neq(column: string, value: any): this {
    this.filters.push({ column, operator: 'neq', value });
    return this;
  }

  public gt(column: string, value: any): this {
    this.filters.push({ column, operator: 'gt', value });
    return this;
  }

  public gte(column: string, value: any): this {
    this.filters.push({ column, operator: 'gte', value });
    return this;
  }

  public lt(column: string, value: any): this {
    this.filters.push({ column, operator: 'lt', value });
    return this;
  }

  public lte(column: string, value: any): this {
    this.filters.push({ column, operator: 'lte', value });
    return this;
  }

  public like(column: string, pattern: string): this {
    this.filters.push({ column, operator: 'like', value: pattern });
    return this;
  }

  public isNull(column: string): this {
    this.filters.push({ column, operator: 'is', value: 'null' });
    return this;
  }

  public isNotNull(column: string): this {
    this.filters.push({ column, operator: 'is', value: 'notnull' });
    return this;
  }

  public in(column: string, values: any[]): this {
    const formatted = `(${values.join(',')})`;
    this.filters.push({ column, operator: 'in', value: formatted });
    return this;
  }

  // Sorting & Pagination
  public order(column: string, direction: 'asc' | 'desc' = 'asc'): this {
    this.orderBy = `${column}.${direction}`;
    return this;
  }

  public limit(value: number): this {
    this.limitVal = value;
    return this;
  }

  public offset(value: number): this {
    this.offsetVal = value;
    return this;
  }

  // Execution
  public async execute(): Promise<T> {
    const queryParams = new URLSearchParams();

    if (this.selectCols) {
      queryParams.set('select', this.selectCols);
    }
    if (this.orderBy) {
      queryParams.set('order', this.orderBy);
    }
    if (this.limitVal !== undefined) {
      queryParams.set('limit', this.limitVal.toString());
    }
    if (this.offsetVal !== undefined) {
      queryParams.set('offset', this.offsetVal.toString());
    }

    for (const filter of this.filters) {
      queryParams.append(filter.column, `${filter.operator}.${filter.value}`);
    }

    const queryString = queryParams.toString();
    const finalUrl = `${this.url}/v1/rest/${this.table}${queryString ? `?${queryString}` : ''}`;

    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
    };

    const session = this.auth.getSession();
    if (session && session.access_token) {
      headers['Authorization'] = `Bearer ${session.access_token}`;
    }

    const requestOptions: RequestInit = {
      method: this.method,
      headers,
    };

    if (this.method === 'POST' || this.method === 'PATCH') {
      requestOptions.body = JSON.stringify(this.bodyData);
    }

    const res = await fetch(finalUrl, requestOptions);
    const data = await res.json();

    if (!res.ok) {
      throw new Error(data.error || `REST request to table ${this.table} failed`);
    }

    return data as T;
  }
}
