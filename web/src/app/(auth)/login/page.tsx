import { LoginForm } from '@/components/auth/login-form';

export default function LoginPage() {
  return (
    <main
      style={{
        minHeight: '100vh',
        display: 'grid',
        placeItems: 'center',
        padding: 24,
        background: 'linear-gradient(180deg, #f8fafc 0%, #eef4fb 100%)',
      }}
    >
      <LoginForm />
    </main>
  );
}

