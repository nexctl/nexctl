export interface ApiEnvelope<T> {
  code: number;
  message: string;
  data: T;
}

export interface OptionItem {
  label: string;
  value: string;
}

