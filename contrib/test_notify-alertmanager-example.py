import unittest
import importlib

loader = importlib.machinery.SourceFileLoader('notify', 'notify-alertmanager-example.py')
mod = loader.load_module()

class TestNotify(unittest.TestCase):

    def test_generate_alert(self):
        args = mod.cli(['--debug', '--alert-api-url', 'https://localhost', '--hostname', 'barfoo', '--state', '1', '--service', 'foobar'])
        actual = mod.generate_alert(args)[0]
        self.assertEqual(actual['labels']['service'], 'foobar')
        self.assertEqual(actual['labels']['instance'], 'barfoo')
        self.assertEqual(actual['labels']['alertname'], 'foobar_barfoo')

if __name__ == '__main__':
    unittest.main()
