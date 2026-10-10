import unittest
from startup_instrumentation import PACKAGE, RUNNER, bluray_component, require_bluray_success


class InstrumentationTests(unittest.TestCase):
    def test_debug_and_separate_components_are_distinct(self):
        for separate in (False, True):
            owner = PACKAGE + '.test' if separate else PACKAGE
            component = owner + '/' + RUNNER
            listing = f'instrumentation:{component} (target={PACKAGE})'
            self.assertEqual(bluray_component(listing, separate), component)
            with self.assertRaises(RuntimeError):
                bluray_component(listing, not separate)

    def test_missing_or_wrong_target_cannot_pass(self):
        for listing in ('', RUNNER,
                        f'instrumentation:{PACKAGE}.test/{RUNNER} (target=another.app)'):
            with self.assertRaises(RuntimeError):
                bluray_component(listing, True)

    def test_pass_marker_requires_successful_completion(self):
        require_bluray_success('MUKSMATT_BLURAY_SAF_PASS\nINSTRUMENTATION_CODE: -1\n')
        for output in ('', 'MUKSMATT_BLURAY_SAF_PASS', 'INSTRUMENTATION_CODE: -1',
                       'MUKSMATT_BLURAY_SAF_PASS\nINSTRUMENTATION_CODE: 0',
                       'MUKSMATT_BLURAY_SAF_PASS\nINSTRUMENTATION_CODE: -1\nINSTRUMENTATION_FAILED: crash'):
            with self.assertRaises(RuntimeError):
                require_bluray_success(output)
