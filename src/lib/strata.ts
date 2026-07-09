import { StrataClient } from '@strata/sdk';

const strataUrl = import.meta.env.VITE_STRATA_URL || 'http://localhost:8000';

export const strata = new StrataClient(strataUrl);
