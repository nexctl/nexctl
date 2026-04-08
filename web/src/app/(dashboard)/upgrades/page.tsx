import { UpgradesPage } from '@/components/upgrades/upgrades-page';
import { getReleases } from '@/services/upgrade';

export default async function UpgradesRoutePage() {
  const releases = await getReleases();
  return <UpgradesPage releases={releases} />;
}
