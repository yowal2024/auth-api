import fs from 'node:fs';
import path from 'node:path';
import {parse,compileTemplate} from '@vue/compiler-sfc';
import * as babel from '@babel/parser';
import traverseModule from '@babel/traverse';
const traverse=traverseModule.default;
function files(dir){return fs.readdirSync(dir,{withFileTypes:true}).flatMap(x=>x.isDirectory()?files(path.join(dir,x.name)):[path.join(dir,x.name)])}
const errors=[];let count=0;
for(const file of files('src').filter(x=>x.endsWith('.vue'))){
 count++;const {descriptor,errors:pe}=parse(fs.readFileSync(file,'utf8'));errors.push(...pe.map(x=>({file,error:String(x)})));
 let compiled;try{compiled=compileTemplate({source:descriptor.template.content,filename:file,id:'check'});}catch(e){errors.push({file,error:e.message});continue}
 errors.push(...compiled.errors.map(x=>({file,error:String(x)})));
 // A controlled tab without an update handler renders but cannot be switched.
 for(const tag of descriptor.template.content.matchAll(/<Tabs\b[^>]*>/g)) {
  if(/:active-key\s*=/.test(tag[0]) && !/v-model:active-key|@(?:update:activeKey|update:active-key|change|tab-click)/.test(tag[0]))
   errors.push({file,error:'Tabs active-key is read-only; bind v-model or handle updates'});
 }
 const bindings=new Set(['$slots','$attrs','$emit','$route','$router','$event','undefined']);
 if(descriptor.script){const ast=babel.parse(descriptor.script.content,{sourceType:'module'});
 traverse(ast,{ExportDefaultDeclaration(p){for(const prop of p.node.declaration.properties||[]){const name=prop.key?.name;if(name==='components'||name==='props')for(const field of prop.value.properties||[])bindings.add(field.key.name||field.key.value);if(name==='setup'){const ret=prop.body.body.findLast(x=>x.type==='ReturnStatement');for(const f of ret?.argument?.properties||[])bindings.add(f.key.name||f.key.value)}}}});
 }
 try{const ast=babel.parse(compiled.code,{sourceType:'module'});const missing=new Set();traverse(ast,{MemberExpression(p){if(p.node.object.name==='_ctx'&&!bindings.has(p.node.property.name))missing.add(p.node.property.name)}});if(missing.size)errors.push({file,missing:[...missing]})}catch{}
}
fs.mkdirSync('recovery',{recursive:true});fs.writeFileSync('recovery/source-check.json',JSON.stringify({components:count,errors},null,2));
console.log(JSON.stringify({components:count,errors},null,2));if(errors.length)process.exitCode=1;
