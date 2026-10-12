import importlib.util,io,json,pathlib,tarfile,tempfile,unittest,hashlib
spec=importlib.util.spec_from_file_location('checkpoint',pathlib.Path(__file__).with_name('checkpoint.py'));c=importlib.util.module_from_spec(spec);spec.loader.exec_module(c)
class ArchiveTests(unittest.TestCase):
 def bundle(self,entries):
  root=pathlib.Path(tempfile.mkdtemp());archive=root/'archive.tar.gz'
  with tarfile.open(archive,'w:gz') as tar:
   for name,data in entries:
    info=tarfile.TarInfo(name);info.size=len(data);tar.addfile(info,io.BytesIO(data))
  dest=root/'extract';dest.mkdir();return archive,dest
 def test_path_escape(self):
  archive,dest=self.bundle([('../escape',b'x')])
  with self.assertRaisesRegex(RuntimeError,'unsafe_archive_path'):c.verify_bundle(archive,dest)
 def test_hash(self):
  manifest={'version':'checkpoint-v1','files':{'database.dump':'invalid'}}
  archive,dest=self.bundle([('database.dump',b'data'),('manifest.json',json.dumps(manifest).encode())])
  with self.assertRaisesRegex(RuntimeError,'checkpoint_hash_mismatch'):c.verify_bundle(archive,dest)
 def test_valid(self):
  manifest={'version':'checkpoint-v1','files':{'database.dump':hashlib.sha256(b'data').hexdigest()}}
  archive,dest=self.bundle([('database.dump',b'data'),('manifest.json',json.dumps(manifest).encode())])
  self.assertEqual(c.verify_bundle(archive,dest),manifest)
if __name__=='__main__':unittest.main()
