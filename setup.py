from setuptools import setup, find_packages

from gerrittest import __version__


setup(
    name="gerrittest",
    version=".".join(map(str, __version__)),
    packages=find_packages("."),
    url="https://github.com/opalmer/gerrittest",
    license="MIT",
    author="Oliver Palmer",
    author_email="oliverpalmer@opalmer.com",
    description="A tool for running Gerrit in docker. Mainly intended for local "
                "testing.",
    entry_points={
        "console_scripts": ["gerrittest = gerrittest.entrypoints:main"]
    },
    classifiers=(
        "Development Status :: 2 - Pre-Alpha",
        "Topic :: Software Development :: Testing",
        "Topic :: Utilities"
    ),
)
