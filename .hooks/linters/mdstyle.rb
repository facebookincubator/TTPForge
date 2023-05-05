################################################################################
# Style file for markdownlint.
#
# https://github.com/markdownlint/markdownlint/blob/master/docs/configuration.md
#
# This file is referenced by the project `.mdlrc`.
################################################################################

#===============================================================================
# Start with all built-in rules.
# https://github.com/markdownlint/markdownlint/blob/master/docs/RULES.md
all

#===============================================================================
# Override default parameters for some built-in rules.
# https://github.com/markdownlint/markdownlint/blob/master/docs/creating_styles.md#parameters

exclude_tag :line_length
# Allow long lines in code blocks and tables
rule 'MD013', line_length: 120, ignore_code_blocks: true, tables: false

#===============================================================================
# Exclude the rules I disagree with.

# IMHO it's easier to read lists like:
# * outmost indent
#   - one indent
#   - second indent
# * Another major bullet
exclude_rule 'MD004' # Unordered list style

exclude_rule 'MD007' # Unordered list indentation

# Ordered lists are fine.
exclude_rule 'MD029'

# The first line doesn't always need to be a top level header.
exclude_rule 'MD041'

# I find it necessary to use '<br/>' to force line breaks.
exclude_rule 'MD033' # Inline HTML

# Using bare URLs is fine.
exclude_rule 'MD034'
