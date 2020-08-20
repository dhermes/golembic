# Configuration file for the Sphinx documentation builder.
#
# This file only contains a selection of the most common options. For a full
# list see the documentation:
# https://www.sphinx-doc.org/en/2.0/usage/configuration.html

# -- Path setup --------------------------------------------------------------

# If extensions (or modules to document with autodoc) are in another directory,
# add these directories to sys.path here. If the directory is relative to the
# documentation root, use os.path.abspath to make it absolute, like shown here.
#
# import sys
# sys.path.insert(0, os.path.abspath('.'))
import os
import subprocess

import sphinx_rtd_theme


# -- Project information -----------------------------------------------------


def is_clean():
    cmd_output = subprocess.check_output(("git", "status", "--short"))
    modified_files = cmd_output.decode("ascii").strip()
    return modified_files == ""


def get_version_full():
    cmd_output = subprocess.check_output(("git", "log", "-1", "--pretty=%H"))
    return cmd_output.decode("ascii").strip()


def get_version(version_full, currently_clean):
    short_version = version_full[:8]
    if not currently_clean:
        return "{}-dirty".format(short_version)

    return short_version


def get_github_version(version_full, currently_clean):
    if not currently_clean:
        return "main"

    return version_full


def get_copyright(version):
    return "2020, Danny Hermes Revision {}".format(version)


project = "golembic"
_currently_clean = is_clean()
_version_full = get_version_full()
version = get_version(_version_full, _currently_clean)
copyright = get_copyright(version)
author = "Danny Hermes"


# -- General configuration ---------------------------------------------------

# Add any Sphinx extension module names here, as strings. They can be
# extensions coming with Sphinx (named 'sphinx.ext.*') or your custom
# ones.
extensions = [
    "autoapi.extension",
    "sphinxcontrib.golangdomain",
]
# $ cd ..
# $ docker run --rm --interactive --tty --volume $(pwd):/var/code --workdir /var/code python:2.7.18-slim-buster /bin/bash

# Configure `sphinx-autoapi`.
autoapi_type = "go"
autoapi_dirs = ["..", "../postgres"]

# Add any paths that contain templates here, relative to this directory.
templates_path = ["_templates"]

# List of patterns, relative to source directory, that match files and
# directories to ignore when looking for source files.
# This pattern also affects html_static_path and html_extra_path.
exclude_patterns = []


# -- Options for HTML output -------------------------------------------------

# The theme to use for HTML and HTML Help pages.  See the documentation for
# a list of builtin themes.
#
html_theme = "sphinx_rtd_theme"
html_theme_path = [sphinx_rtd_theme.get_html_theme_path()]

# Add any paths that contain custom static files (such as style sheets) here,
# relative to this directory. They are copied after the builtin static files,
# so a file named "default.css" will overwrite the builtin "default.css".
html_static_path = []

html_context = {
    "display_github": True,
    "github_host": "github.com",
    "github_user": "dhermes",
    "github_repo": "golembic",
    "github_version": get_github_version(_version_full, _currently_clean),
    "conf_py_path": "/docs/",
}
