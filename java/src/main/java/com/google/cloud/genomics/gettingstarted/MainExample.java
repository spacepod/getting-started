/*
Copyright 2014 Google Inc. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package com.google.cloud.genomics.gettingstarted;

import com.beust.jcommander.JCommander;
import com.beust.jcommander.Parameter;
import com.beust.jcommander.ParameterException;
import com.beust.jcommander.internal.Maps;
import com.google.api.client.extensions.java6.auth.oauth2.VerificationCodeReceiver;
import com.google.api.client.extensions.jetty.auth.oauth2.LocalServerReceiver;
import com.google.api.client.googleapis.extensions.java6.auth.oauth2.GooglePromptReceiver;
import com.google.api.services.genomics.Genomics;
import com.google.api.services.genomics.GenomicsScopes;
import com.google.api.services.genomics.model.CallSet;
import com.google.api.services.genomics.model.Read;
import com.google.api.services.genomics.model.Readset;
import com.google.api.services.genomics.model.SearchCallSetsRequest;
import com.google.api.services.genomics.model.SearchReadsRequest;
import com.google.api.services.genomics.model.SearchReadsetsRequest;
import com.google.api.services.genomics.model.SearchVariantsRequest;
import com.google.api.services.genomics.model.Variant;
import com.google.cloud.genomics.utils.GenomicsFactory;
import com.google.common.base.Joiner;
import com.google.common.base.Suppliers;
import com.google.common.collect.Lists;

import java.io.File;
import java.io.IOException;
import java.math.BigInteger;
import java.util.List;
import java.util.Map;

/**
 * Getting started with Google Genomics in Java
 */
public class MainExample {

  private static class Arguments {
    @Parameter(names = "--client_secrets_filename",
        description = "Path to client_secrets.json")
    public String clientSecretsFilename = "client_secrets.json";

    @Parameter(names = "--noauth_local_webserver",
        description = "If your browser is on a different machine then run this" +
            "application with this command-line parameter")
    public boolean noLocalServer = false;
  }

  public static void main(String[] args) throws IOException {
    Arguments arguments = new Arguments();
    JCommander parser = new JCommander(arguments);

    try {
      // Parse the command line
      parser.parse(args);

      // Authorization
      VerificationCodeReceiver receiver = arguments.noLocalServer ?
          new GooglePromptReceiver() : new LocalServerReceiver();
      GenomicsFactory genomicsFactory = GenomicsFactory.builder("getting_started_java")
          .setScopes(Lists.newArrayList(GenomicsScopes.GENOMICS))
          .setVerificationCodeReceiver(Suppliers.ofInstance(receiver))
          .build();

      File clientSecrets = new File(arguments.clientSecretsFilename);
      if (!clientSecrets.exists()) {
        System.err.println(
            "Client secrets file " + arguments.clientSecretsFilename + " does not exist."
            + " Visit https://developers.google.com/genomics to learn how"
            + " to install a client_secrets.json file.  If you have installed a client_secrets.json"
            + " in a specific location, use --client_secrets_filename <path>/client_secrets.json.");
        return;
      }
      Genomics genomics = genomicsFactory.fromClientSecretsFile(clientSecrets);


      //
      // This example gets the read bases for a sample at specific a position
      //
      String datasetId = "376902546192"; // This is the 1000 Genomes dataset ID
      String sample = "NA12872";
      String referenceName = "22";
      final Integer referencePosition = 51003836;

      // 1. First find the readset ID for the sample
      SearchReadsetsRequest readsetsReq = new SearchReadsetsRequest()
          .setDatasetIds(Lists.newArrayList(datasetId))
          .setName(sample);

      List<Readset> readsets = genomics.readsets().search(readsetsReq)
          .setFields("readsets(id)").execute().getReadsets();
      if (readsets == null || readsets.size() != 1) {
        System.err.println("Searching for " + sample
            + " didn't return the right number of readsets");
        return;
      }

      String readsetId = readsets.get(0).getId();


      // 2. Once we have the readset ID,
      // lookup the reads at the position we are interested in
      SearchReadsRequest readsReq = new SearchReadsRequest()
          .setReadsetIds(Lists.newArrayList(readsetId))
          .setSequenceName(referenceName)
          .setSequenceStart(BigInteger.valueOf(referencePosition))
          .setSequenceEnd(BigInteger.valueOf(referencePosition))
          .setMaxResults(BigInteger.valueOf(1024));

      List<Read> reads = genomics.reads().search(readsReq)
          .setFields("reads(position,originalBases,cigar)").execute().getReads();

      Map<Character, Integer> baseCounts = Maps.newHashMap();
      for (Read read : reads) {
        int index = referencePosition - read.getPosition();
        // Note: This is simplistic - the cigar should be considered for real code
        Character base = read.getOriginalBases().charAt(index);

        if (!baseCounts.containsKey(base)) {
          baseCounts.put(base, 0);
        }
        baseCounts.put(base, baseCounts.get(base) + 1);
      }

      System.out.println(sample + " bases on " + referenceName + " at "
          + referencePosition + " are");
      for (Map.Entry<Character, Integer> entry : baseCounts.entrySet()) {
        System.out.println(entry.getKey() + ": " + entry.getValue());
      }


      //
      // This example gets the variants for a sample at a specific position
      //

      // 1. First find the call set ID for the sample
      SearchCallSetsRequest callSetsReq = new SearchCallSetsRequest()
          .setVariantSetIds(Lists.newArrayList(datasetId))
          .setName(sample);

      List<CallSet> callSets = genomics.callsets().search(callSetsReq)
          .setFields("callSets(id)").execute().getCallSets();
      if (callSets == null || callSets.size() != 1) {
        System.err.println("Searching for " + sample
            + " didn't return the right number of call sets");
        return;
      }

      String callSetId = callSets.get(0).getId();


      // 2. Once we have the call set ID,
      // lookup the variants that overlap the position we are interested in
      SearchVariantsRequest variantsReq = new SearchVariantsRequest()
          .setCallSetIds(Lists.newArrayList(callSetId))
          .setReferenceName(referenceName)
          // Note: currently, variants are 0-based and reads are 1-based,
          // reads will move to 0-based coordinates in the next version of the API
          .setStart(Long.valueOf(referencePosition) - 1)
          .setEnd(Long.valueOf(referencePosition));

      Variant variant = genomics.variants().search(variantsReq)
          .setFields("variants(names,referenceBases,alternateBases,calls(genotype))")
          .execute().getVariants().get(0);

      String variantName = variant.getNames().get(0);

      List<String> genotype = Lists.newArrayList();
      for (Integer g : variant.getCalls().get(0).getGenotype()) {
        if (g == 0) {
          genotype.add(variant.getReferenceBases());
        } else {
          genotype.add(variant.getAlternateBases().get(g - 1));
        }
      }

      System.out.println("the called genotype is " + Joiner.on(',').join(genotype)
          + " at " + variantName);


    } catch (ParameterException e) {
      System.err.append(e.getMessage()).append("\n");
      parser.usage();
    } catch (IllegalStateException e) {
      System.err.println(e.getMessage());
    } catch (Throwable t) {
      t.printStackTrace();
    }
  }
}
