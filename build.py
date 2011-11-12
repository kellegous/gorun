#!/usr/bin/env python

import optparse
import os
import subprocess
import sys

def Test(here, goroot, builddir):
  def LoadTests():
    with open(os.path.join(here, 'tests', 'tests')) as f:
      for line in f.xreadlines():
        name, args = line.split(':', 1)
        shouldPass = name[0] == '+'
        yield (name[1:].strip(), args.strip(), shouldPass)
  def Run(name, testdir, builddir, args):
    p = subprocess.Popen([
        os.path.abspath(os.path.join(builddir, 'gorun')),
        '--build-dir=%s' % os.path.abspath(os.path.join(builddir, 'tests', name))
      ] + args, cwd = testdir, stdout = subprocess.PIPE)
    out, err = p.communicate()
    return p.returncode == 0
  for name, args, shouldPass in LoadTests():
    didPass = Run(name, os.path.join(here, 'tests'), builddir, args.split(' '))
    print "%s | %s" % (name.ljust(30), shouldPass == didPass)

def Build(here, goroot, builddir):
  if not os.path.exists(builddir):
    os.makedirs(builddir)
  if subprocess.call([
      os.path.join(goroot, 'bin', '6g'),
      '-o',
      os.path.join(builddir, 'gorun.6'),
      os.path.join(here, 'gorun.go')]) != 0:
    return False
  if subprocess.call([
      os.path.join(goroot, 'bin', '6l'),
      '-o',
      os.path.join(builddir, 'gorun'),
      os.path.join(builddir, 'gorun.6')]) != 0:
    return False
  return True
    
if __name__ == '__main__':
  here = os.path.dirname(__file__)
  parser = optparse.OptionParser()
  parser.add_option('--goroot',
      dest = 'goroot',
      default = os.path.expanduser('~/src/go'),
      help = '')
  parser.add_option('--build-dir',
      dest = 'builddir',
      default = os.path.join(here, 'bin'),
      help = '')
  options, args = parser.parse_args()

  env = os.environ
  env['PATH'] = '%s:%s' % (options.goroot, env['PATH'])
  env['GOROOT'] = options.goroot

  if not Build(here, options.goroot, options.builddir):
    sys.exit(1)

  while len(args) > 0:
    if args[0] == 'test' and not Test(here, options.goroot, options.builddir):
      sys.exit(1)
    args = args[1:]

  sys.exit(0)
