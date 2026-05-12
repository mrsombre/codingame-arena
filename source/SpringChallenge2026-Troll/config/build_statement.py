#! /usr/bin/env python
import re

def fill_template(template, target):
    with open('../src/main/java/engine/Constants.java') as f:
        constants = f.read()
    var_values = []
    for line in constants.split('\n'):
        if not '=' in line: continue
        assign = line.split('=')
        left = assign[0].strip().split()[-1]
        right = assign[1].strip()
        right = right[0:right.index(';')]
        if '{' in right:
            right = right.replace('{', '').replace('}', '').split(',')
            for i in range(0, len(right)):
                var_values.append([left+'_'+str(i), right[i].strip()])
        else:
            var_values.append([left, right])

    with open(template) as f:
        statement = f.read()

    for rep in sorted(var_values, key = lambda v: -len(v[0])):
        statement = statement.replace(rep[0], rep[1])


    pattern = re.compile(r"\[\[([^[]+)\]\]")

    def replace_placeholders(text):
        def repl(match):
            expr = match.group(1)
            return '<const>' + str(eval(expr)) + '</const>'
        return pattern.sub(repl, text)

    statement = replace_placeholders(statement)

    with open(target, 'w') as f:
        f.write(statement)

fill_template('statement_en.html.param', 'statement_en.html.tpl')
fill_template('statement_fr.html.param', 'statement_fr.html.tpl')