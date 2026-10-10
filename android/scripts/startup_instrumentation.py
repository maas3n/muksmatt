"""Resolve only instrumentation belonging to the expected APK and target."""
import re

PACKAGE = "io.github.maas3n.muksmatt"
RUNNER = "io.github.maas3n.mattmux.BluraySmokeInstrumentation"


def bluray_component(listing, separate=False):
    owner = PACKAGE + ".test" if separate else PACKAGE
    component = owner + "/" + RUNNER
    expected = f"instrumentation:{component} (target={PACKAGE})"
    if expected not in [line.strip() for line in listing.splitlines()]:
        raise RuntimeError("Required Blu-ray SAF device test instrumentation was not packaged: " + component)
    return component


def require_bluray_success(output):
    if ("MUKSMATT_BLURAY_SAF_PASS" not in output
            or "INSTRUMENTATION_CODE: -1" not in output
            or re.search(r"INSTRUMENTATION_(?:FAILED|ABORTED)|FAILURES!!!", output)):
        raise RuntimeError("Android Blu-ray ISO/BDMV SAF remux failed: " + output)
