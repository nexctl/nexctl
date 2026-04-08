import { FilesPage } from '@/components/files/files-page';
import { getFiles } from '@/services/file';

export default async function FilesRoutePage() {
  const files = await getFiles();
  return <FilesPage files={files} />;
}
